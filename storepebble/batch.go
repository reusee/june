// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storepebble

import (
	"context"
	"fmt"
	"io"
	"sync"
	"sync/atomic"

	"github.com/cockroachdb/pebble"
	"github.com/reusee/june/index"
	"github.com/reusee/june/storekv"
	"github.com/reusee/pr"
	"github.com/reusee/pr2"
)

type Batch struct {
	*pr2.WaitGroup
	ctx      context.Context
	name     string
	store    *Store
	batch    *pebble.Batch
	cond     *sync.Cond
	numRead  int
	numWrite int
}

var batchSerial int64

type NewBatch func(
	wt *pr.WaitTree,
	store *Store,
) (*Batch, error)

func (Def) NewBatch() NewBatch {
	return func(
		parentWaitTree *pr.WaitTree,
		store *Store,
	) (_ *Batch, err error) {
		defer he(&err)
		batch := store.DB.NewIndexedBatch()
		b := &Batch{
			name: fmt.Sprintf("pebble-batch%d(%s)",
				atomic.AddInt64(&batchSerial, 1),
				store.Name(),
			),
			store: store,
			batch: batch,
			cond:  sync.NewCond(new(sync.Mutex)),
		}
		b.ctx, b.WaitGroup = pr2.NewWaitGroup(parentWaitTree.Ctx)
		parentWaitTree.Go(func() {
			<-parentWaitTree.Ctx.Done()
			b.Wait()
			ce(batch.Close())
		})
		return b, nil
	}
}

func (b *Batch) Name() string {
	return b.name
}

func (b *Batch) StoreID() string {
	return b.store.StoreID()
}

func (b *Batch) lockRead() func() {
	b.cond.L.Lock()
	defer b.cond.L.Unlock()
	for b.numWrite > 0 {
		b.cond.Wait()
	}
	b.numRead++
	return func() {
		b.cond.L.Lock()
		defer b.cond.L.Unlock()
		b.numRead--
		b.cond.Broadcast()
	}
}

func (b *Batch) lockWrite() func() {
	b.cond.L.Lock()
	defer b.cond.L.Unlock()
	for b.numRead > 0 || b.numWrite > 0 {
		b.cond.Wait()
	}
	b.numWrite++
	return func() {
		b.cond.L.Lock()
		defer b.cond.L.Unlock()
		b.numWrite--
		b.cond.Broadcast()
	}
}

func (b *Batch) Commit() (err error) {
	select {
	case <-b.ctx.Done():
		return b.ctx.Err()
	default:
	}
	defer b.Add()()
	defer b.lockWrite()()
	return b.batch.Commit(writeOptions)
}

func (b *Batch) Abort() error {
	select {
	case <-b.ctx.Done():
		return b.ctx.Err()
	default:
	}
	defer b.Add()()
	defer b.lockWrite()()
	return b.batch.Close()
}

var _ storekv.KV = new(Batch)

func (b *Batch) CostInfo() storekv.CostInfo {
	return costInfo
}

func (b *Batch) KeyDelete(keys ...string) (err error) {
	select {
	case <-b.ctx.Done():
		return b.ctx.Err()
	default:
	}
	defer b.Add()()
	return b.store.keyDelete(b.Add, b.delete, keys...)
}

func (b *Batch) delete(
	key []byte,
	options *pebble.WriteOptions,
) error {
	select {
	case <-b.ctx.Done():
		return b.ctx.Err()
	default:
	}
	defer b.lockWrite()()
	if err := b.batch.Delete(key, writeOptions); err != nil {
		return err
	}
	return nil
}

func (b *Batch) KeyExists(key string) (ok bool, err error) {
	select {
	case <-b.ctx.Done():
		return false, b.ctx.Err()
	default:
	}
	defer b.Add()()
	return b.store.keyExists(b.Add, b.get, key)
}

func (b *Batch) get(
	key []byte,
) (
	[]byte,
	io.Closer,
	error,
) {
	select {
	case <-b.ctx.Done():
		return nil, nil, b.ctx.Err()
	default:
	}
	defer b.lockRead()()
	return b.batch.Get(key)
}

func (b *Batch) KeyGet(key string, fn func(io.Reader) error) (err error) {
	select {
	case <-b.ctx.Done():
		return b.ctx.Err()
	default:
	}
	defer b.Add()()
	return b.store.keyGet(b.Add, b.get, key, fn)
}

func (b *Batch) KeyIter(prefix string, fn func(key string) error) (err error) {
	select {
	case <-b.ctx.Done():
		return b.ctx.Err()
	default:
	}
	defer b.Add()()
	return b.store.keyIter(
		b.Add,
		b.newIter,
		prefix,
		func(fn func()) {
			defer b.lockRead()()
			fn()
		},
		fn,
	)
}

func (b *Batch) newIter(options *pebble.IterOptions) *pebble.Iterator {
	defer b.lockRead()()
	return b.batch.NewIter(options)
}

func (b *Batch) KeyPut(key string, r io.Reader) (err error) {
	select {
	case <-b.ctx.Done():
		return b.ctx.Err()
	default:
	}
	defer b.Add()()
	return b.store.keyPut(b.Add, b.get, b.set, key, r)
}

func (b *Batch) set(
	key []byte,
	value []byte,
	option *pebble.WriteOptions,
) (err error) {
	select {
	case <-b.ctx.Done():
		return b.ctx.Err()
	default:
	}
	defer b.lockWrite()()
	return b.batch.Set(key, value, option)
}

var _ index.IndexManager = new(Batch)

func (b *Batch) IndexFor(id StoreID) (index.Index, error) {
	select {
	case <-b.ctx.Done():
		return nil, b.ctx.Err()
	default:
	}
	defer b.Add()()
	return Index{
		ctx: b.ctx,
		name: fmt.Sprintf("pebble-batch-index%d(%v, %v)",
			atomic.AddInt64(&indexSerial, 1),
			b.Name(),
			id,
		),
		begin: b.Add,
		exists: func(key []byte) (bool, error) {
			_, cl, err := b.get(key)
			if is(err, pebble.ErrNotFound) {
				return false, nil
			}
			if err != nil {
				return false, err
			}
			cl.Close()
			return true, nil
		},
		set: func(key []byte, value []byte, options *pebble.WriteOptions) error {
			return b.set(key, value, options)
		},
		delete: func(key []byte, options *pebble.WriteOptions) error {
			return b.delete(key, options)
		},
		newIter: func(options *pebble.IterOptions) *pebble.Iterator {
			return b.newIter(options)
		},
		withRLock: func(fn func()) {
			defer b.lockRead()()
			fn()
		},
		id: id,
	}, nil
}

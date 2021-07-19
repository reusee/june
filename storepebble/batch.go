// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storepebble

import (
	"fmt"
	"io"
	"sync"
	"sync/atomic"

	"github.com/cockroachdb/pebble"
	"github.com/reusee/june/index"
	"github.com/reusee/june/storekv"
	"github.com/reusee/pr"
)

type Batch struct {
	*pr.WaitTree

	name string

	store *Store

	batch *pebble.Batch

	l sync.RWMutex

	closeOnce sync.Once
}

var batchSerial int64

type NewBatch func(
	store *Store,
) (*Batch, error)

func (_ Def) NewBatch(
	parentWaitTree *pr.WaitTree,
) NewBatch {
	return func(
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
		}
		b.WaitTree = pr.NewWaitTree(parentWaitTree)
		parentWaitTree.Go(func() {
			<-parentWaitTree.Ctx.Done()
			b.WaitTree.Wait()
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

func (b *Batch) Commit() (err error) {
	defer he(&err)
	b.l.Lock()
	defer b.l.Unlock()
	ce(b.batch.Commit(writeOptions))
	return
}

func (b *Batch) Abort() error {
	return b.batch.Close()
}

var _ storekv.KV = new(Batch)

func (b *Batch) KeyDelete(keys ...string) (err error) {
	select {
	case <-b.Ctx.Done():
		return b.Ctx.Err()
	default:
	}
	defer b.Add()()
	b.l.Lock()
	defer b.l.Unlock()
	return b.store.keyDelete(b.Add, b.delete, keys...)
}

func (b *Batch) delete(
	key []byte,
	options *pebble.WriteOptions,
) error {
	select {
	case <-b.Ctx.Done():
		return b.Ctx.Err()
	default:
	}
	b.l.Lock()
	defer b.l.Unlock()
	if err := b.batch.Delete(key, writeOptions); err != nil {
		return err
	}
	return nil
}

func (b *Batch) KeyExists(key string) (ok bool, err error) {
	select {
	case <-b.Ctx.Done():
		return false, b.Ctx.Err()
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
	case <-b.Ctx.Done():
		return nil, nil, b.Ctx.Err()
	default:
	}
	b.l.RLock()
	defer b.l.RUnlock()
	return b.batch.Get(key)
}

func (b *Batch) KeyGet(key string, fn func(io.Reader) error) (err error) {
	select {
	case <-b.Ctx.Done():
		return b.Ctx.Err()
	default:
	}
	defer b.Add()()
	return b.store.keyGet(b.Add, b.get, key, fn)
}

func (b *Batch) KeyIter(prefix string, fn func(key string) error) (err error) {
	select {
	case <-b.Ctx.Done():
		return b.Ctx.Err()
	default:
	}
	defer b.Add()()
	return b.store.keyIter(
		b.Add,
		b.batch.NewIter,
		prefix,
		func(fn func()) {
			b.l.RLock()
			defer b.l.RUnlock()
			fn()
		},
		fn,
	)
}

func (b *Batch) KeyPut(key string, r io.Reader) (err error) {
	select {
	case <-b.Ctx.Done():
		return b.Ctx.Err()
	default:
	}
	defer b.Add()()
	b.l.Lock()
	defer b.l.Unlock()
	return b.store.keyPut(b.Add, b.get, b.set, key, r)
}

func (b *Batch) set(
	key []byte,
	value []byte,
	option *pebble.WriteOptions,
) (err error) {
	select {
	case <-b.Ctx.Done():
		return b.Ctx.Err()
	default:
	}
	b.l.Lock()
	defer b.l.Unlock()
	return b.batch.Set(key, value, option)
}

var _ index.IndexManager = new(Batch)

func (b *Batch) IndexFor(id StoreID) (index.Index, error) {
	select {
	case <-b.Ctx.Done():
		return nil, b.Ctx.Err()
	default:
	}
	defer b.Add()()
	return Index{
		ctx: b.Ctx,
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
			b.l.Lock()
			defer b.l.Unlock()
			return b.set(key, value, options)
		},
		delete: func(key []byte, options *pebble.WriteOptions) error {
			b.l.Lock()
			defer b.l.Unlock()
			return b.delete(key, options)
		},
		newIter: func(options *pebble.IterOptions) *pebble.Iterator {
			return b.batch.NewIter(options)
		},
		withRLock: func(fn func()) {
			b.l.RLock()
			defer b.l.RUnlock()
			fn()
		},
		id: id,
	}, nil
}

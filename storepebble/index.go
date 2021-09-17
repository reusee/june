// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storepebble

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"
	"sync/atomic"

	"github.com/cockroachdb/pebble"
	"github.com/reusee/e4"
	"github.com/reusee/june/index"
	"github.com/reusee/sb"
)

var _ index.IndexManager = new(Store)

func (s *Store) IndexFor(id StoreID) (index.Index, error) {
	defer s.Add()()
	return Index{
		ctx: s.Ctx,
		name: fmt.Sprintf("pebble-index%d(%v, %v)",
			atomic.AddInt64(&indexSerial, 1),
			s.Name(),
			id,
		),
		begin: s.Add,
		exists: func(key []byte) (bool, error) {
			_, cl, err := s.DB.Get(key)
			if is(err, pebble.ErrNotFound) {
				return false, nil
			}
			if err != nil {
				return false, err
			}
			cl.Close()
			return true, nil
		},
		set:    s.DB.Set,
		delete: s.DB.Delete,
		newIter: func(options *pebble.IterOptions) *pebble.Iterator {
			return s.DB.NewIter(options)
		},
		withRLock: func(fn func()) {
			fn()
		},
		id: id,
	}, nil
}

var indexSerial int64

type Index struct {
	ctx     context.Context
	name    string
	begin   func() func()
	exists  func([]byte) (bool, error)
	set     func([]byte, []byte, *pebble.WriteOptions) error
	delete  func([]byte, *pebble.WriteOptions) error
	newIter func(*pebble.IterOptions) *pebble.Iterator

	withRLock func(func())
	id        StoreID
}

var indexBufPool = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

func (i Index) Name() string {
	return i.name
}

func (i Index) Save(entry IndexEntry, options ...IndexSaveOption) (err error) {
	select {
	case <-i.ctx.Done():
		return i.ctx.Err()
	default:
	}
	defer catchErr(&err, pebble.ErrClosed)
	defer i.begin()()

	if entry.Type == nil {
		return we.With(
			e4.NewInfo("entry type is nil: %+v", entry),
		)(index.ErrInvalidEntry)
	}
	if entry.Key == nil && entry.Path == nil {
		return we.With(
			e4.NewInfo("both entry key and path is nil: %+v", entry),
		)(index.ErrInvalidEntry)
	}

	var tapEntry []IndexTapEntry
	for _, option := range options {
		switch option := option.(type) {
		case IndexTapEntry:
			tapEntry = append(tapEntry, option)
		default:
			panic(fmt.Errorf("unknown option: %T", option))
		}
	}

	if err := i.save(entry); err != nil {
		return err
	}
	if entry.Key != nil {
		if err := i.save(index.PreEntry{
			Key:   *entry.Key,
			Type:  entry.Type,
			Tuple: entry.Tuple,
		}); err != nil {
			return err
		}
	}

	for _, tap := range tapEntry {
		tap(entry)
	}

	return nil
}

func (i Index) save(obj any) error {
	buf := indexBufPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		indexBufPool.Put(buf)
	}()
	if err := sb.Copy(
		sb.Marshal(sb.Tuple{
			Idx,
			StoreIndex{
				ID:    i.id,
				Value: sb.Marshal(obj),
			},
		}),
		sb.Encode(buf),
	); err != nil {
		return err
	}
	bs := buf.Bytes()
	yes, err := i.exists(bs)
	if err != nil {
		return err
	}
	if yes {
		return nil
	}
	if err = i.set(bs, []byte{}, writeOptions); err != nil {
		return err
	}
	return nil
}

func (i Index) Delete(entry IndexEntry) (err error) {
	select {
	case <-i.ctx.Done():
		return i.ctx.Err()
	default:
	}
	defer catchErr(&err, pebble.ErrClosed)
	defer i.begin()()

	if err := i._delete(entry); err != nil {
		return err
	}
	if entry.Key != nil {
		if err := i._delete(index.PreEntry{
			Key:   *entry.Key,
			Type:  entry.Type,
			Tuple: entry.Tuple,
		}); err != nil {
			return err
		}
	}

	return nil
}

func (i Index) _delete(obj any) error {
	buf := indexBufPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		indexBufPool.Put(buf)
	}()
	if err := sb.Copy(
		sb.Marshal(sb.Tuple{
			Idx,
			StoreIndex{
				ID:    i.id,
				Value: sb.Marshal(obj),
			},
		}),
		sb.Encode(buf),
	); err != nil {
		return err
	}
	if err := i.delete(buf.Bytes(), writeOptions); err != nil {
		return err
	}
	return nil
}

type indexIter struct {
	end       func()
	iter      *pebble.Iterator
	bufs      []*bytes.Buffer
	closeOnce sync.Once
	order     index.Order
	ok        bool
	withRLock func(func())
}

func (idx Index) Iter(
	lower *sb.Tokens,
	upper *sb.Tokens,
	order index.Order,
) (ProcSrc, io.Closer, error) {
	select {
	case <-idx.ctx.Done():
		return nil, nil, idx.ctx.Err()
	default:
	}

	iterOptions := new(pebble.IterOptions)
	i := &indexIter{
		order: order,
	}

	if lower != nil {
		buf := indexBufPool.Get().(*bytes.Buffer)
		i.bufs = append(i.bufs, buf)
		if err := sb.Copy(
			sb.Marshal(sb.Tuple{
				Idx,
				StoreIndex{
					ID:    idx.id,
					Value: lower.Iter(),
				},
			}),
			sb.Encode(buf),
		); err != nil {
			return nil, nil, err
		}
		iterOptions.LowerBound = buf.Bytes()
	} else {
		buf := indexBufPool.Get().(*bytes.Buffer)
		i.bufs = append(i.bufs, buf)
		if err := sb.Copy(
			sb.Marshal(sb.Tuple{
				Idx,
				StoreIndex{
					ID:    idx.id,
					Value: sb.Marshal(sb.Min),
				},
			}),
			sb.Encode(buf),
		); err != nil {
			return nil, nil, err
		}
		iterOptions.LowerBound = buf.Bytes()
	}

	if upper != nil {
		buf := indexBufPool.Get().(*bytes.Buffer)
		i.bufs = append(i.bufs, buf)
		if err := sb.Copy(
			sb.Marshal(sb.Tuple{
				Idx,
				StoreIndex{
					ID:    idx.id,
					Value: upper.Iter(),
				},
			}),
			sb.Encode(buf),
		); err != nil {
			return nil, nil, err
		}
		iterOptions.UpperBound = buf.Bytes()
	} else {
		buf := indexBufPool.Get().(*bytes.Buffer)
		i.bufs = append(i.bufs, buf)
		if err := sb.Copy(
			sb.Marshal(sb.Tuple{
				Idx,
				StoreIndex{
					ID:    idx.id,
					Value: sb.Marshal(sb.Max),
				},
			}),
			sb.Encode(buf),
		); err != nil {
			return nil, nil, err
		}
		iterOptions.UpperBound = buf.Bytes()
	}

	var ok bool
	var iter *pebble.Iterator
	idx.withRLock(func() {
		iter = idx.newIter(iterOptions)
		if order == index.Asc {
			ok = iter.First()
		} else {
			ok = iter.Last()
		}
	})
	i.iter = iter
	i.ok = ok
	i.end = idx.begin()
	i.withRLock = idx.withRLock

	return i.Iter, i, nil
}

func (p *indexIter) Close() (err error) {
	defer he(&err)
	defer p.end()
	p.closeOnce.Do(func() {
		for _, buf := range p.bufs {
			buf.Reset()
			indexBufPool.Put(buf)
		}
		p.withRLock(func() {
			err = p.iter.Error()
			ce(err, e4.Close(p.iter))
			err = p.iter.Close()
			ce(err)
		})
	})
	return
}

func (p *indexIter) Iter() (*sb.Proc, ProcSrc, error) {
	if !p.ok {
		return nil, nil, nil
	}
	var tokens sb.Tokens
	tuple := sb.Tuple{
		nil, // Idx
		sb.Tuple{
			nil, // id
			sb.CollectValueTokens(&tokens),
		},
	}
	var bs []byte
	p.withRLock(func() {
		bs = p.iter.Key()
	})
	if err := sb.Copy(
		sb.Decode(bytes.NewReader(bs)),
		sb.Unmarshal(&tuple),
	); err != nil {
		return nil, nil, err
	}
	p.withRLock(func() {
		if p.order == index.Asc {
			p.ok = p.iter.Next()
		} else {
			p.ok = p.iter.Prev()
		}
	})
	proc := tokens.Iter()
	return &proc, p.Iter, nil
}

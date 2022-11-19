// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storepebble

import (
	"bytes"
	"context"
	"io"
	"strings"

	"github.com/cockroachdb/pebble"
	"github.com/reusee/e5"
	"github.com/reusee/june/storekv"
	"github.com/reusee/sb"
)

var _ storekv.KV = new(Store)

func (s *Store) CostInfo() storekv.CostInfo {
	return costInfo
}

func (s *Store) KeyExists(ctx context.Context, key string) (ok bool, err error) {
	return s.keyExists(
		ctx,
		s.Add,
		func(_ context.Context, key []byte) ([]byte, io.Closer, error) {
			return s.DB.Get(key)
		},
		key,
	)
}

func (s *Store) keyExists(
	ctx context.Context,
	add func() func(),
	get func(context.Context, []byte) ([]byte, io.Closer, error),
	key string,
) (ok bool, err error) {
	defer he(&err)
	defer add()()
	defer catchErr(&err, pebble.ErrClosed)
	var c io.Closer
	withMarshalKey(func(key []byte) {
		_, c, err = get(ctx, key)
	}, Kv, key)
	if is(err, pebble.ErrNotFound) {
		return false, nil
	}
	ce(err)
	c.Close()
	return true, nil
}

func (s *Store) KeyGet(ctx context.Context, key string, fn func(io.Reader) error) (err error) {
	return s.keyGet(
		ctx,
		s.Add,
		func(_ context.Context, key []byte) ([]byte, io.Closer, error) {
			return s.DB.Get(key)
		},
		key,
		fn,
	)
}

func (s *Store) keyGet(
	ctx context.Context,
	add func() func(),
	get func(context.Context, []byte) ([]byte, io.Closer, error),
	key string,
	fn func(io.Reader) error,
) (err error) {
	defer he(&err,
		e5.With(storekv.StringKey(key)),
	)
	defer add()()
	defer catchErr(&err, pebble.ErrClosed)

	var c io.Closer
	var bs []byte
	withMarshalKey(func(key []byte) {
		bs, c, err = get(ctx, key)
	}, Kv, key)
	if is(err, pebble.ErrNotFound) {
		return we.With(e5.With(storekv.StringKey(key)))(ErrKeyNotFound)
	}
	ce(err)
	defer c.Close()
	if fn != nil {
		err := fn(bytes.NewReader(bs))
		ce(err)
	}
	return nil
}

func (s *Store) KeyPut(ctx context.Context, key string, r io.Reader) (err error) {
	return s.keyPut(
		ctx,
		s.Add,
		func(_ context.Context, key []byte) ([]byte, io.Closer, error) {
			return s.DB.Get(key)
		},
		func(_ context.Context, key, value []byte, opts *pebble.WriteOptions) error {
			return s.DB.Set(key, value, opts)
		},
		key,
		r,
	)
}

func (s *Store) keyPut(
	ctx context.Context,
	add func() func(),
	get func(context.Context, []byte) ([]byte, io.Closer, error),
	set func(context.Context, []byte, []byte, *pebble.WriteOptions) error,
	key string,
	r io.Reader,
) (err error) {
	defer he(&err,
		e5.With(storekv.StringKey(key)),
	)
	defer add()()
	defer catchErr(&err, pebble.ErrClosed)

	var c io.Closer
	var bsKey []byte
	withMarshalKey(func(k []byte) {
		bsKey = append(k[:0:0], k...)
	}, Kv, key)
	_, c, err = get(ctx, bsKey)
	if err == nil {
		c.Close()
		return nil
	}
	var bs []byte
	if b, ok := r.(interface {
		Bytes() []byte
	}); ok {
		bs = b.Bytes()
	} else {
		bs, err = io.ReadAll(r)
		ce(err)
	}
	ce(set(ctx, bsKey, bs, writeOptions))
	return nil
}

func (s *Store) KeyIter(ctx context.Context, prefix string, fn func(key string) error) (err error) {
	return s.keyIter(
		s.Add,
		s.DB.NewIter,
		prefix,
		func(fn func()) {
			fn()
		},
		fn,
	)
}

func (s *Store) keyIter(
	add func() func(),
	newIter func(*pebble.IterOptions) *pebble.Iterator,
	prefix string,
	withRLock func(func()),
	fn func(key string) error,
) (err error) {
	defer he(&err, e5.Info("prefix %s", prefix))
	defer add()()

	var lowerBytes, upperBytes []byte
	withMarshalKey(func(prefix []byte) {
		lowerBytes = append(prefix[:0:0], prefix...)
	}, Kv, prefix)
	withMarshalKey(func(max []byte) {
		upperBytes = append(max[:0:0], max...)
	}, Kv, sb.Max)
	var iter *pebble.Iterator
	withRLock(func() {
		iter = newIter(&pebble.IterOptions{
			LowerBound: lowerBytes,
			UpperBound: upperBytes,
		})
	})

	defer func() {
		withRLock(func() {
			if e := iter.Error(); e != nil {
				err = e
			}
			if e := iter.Close(); e != nil {
				err = e
			}
		})
	}()

	var ok bool
	withRLock(func() {
		ok = iter.First()
	})
	if !ok {
		return nil
	}

	for {
		var bs []byte
		withRLock(func() {
			bs = iter.Key()
		})
		var key string
		if err := sb.Copy(
			sb.Decode(bytes.NewReader(bs)),
			sb.Unmarshal(func(_ Prefix, k string) {
				key = k
			}),
		); err != nil {
			return err
		}
		if !strings.HasPrefix(key, prefix) {
			break
		}
		err = fn(key)
		if is(err, Break) {
			return nil
		}
		ce(err, e5.Info("key %s", key))
		withRLock(func() {
			ok = iter.Next()
		})
		if !ok {
			return nil
		}
	}

	return nil
}

func (s *Store) KeyDelete(ctx context.Context, keys ...string) (err error) {
	return s.keyDelete(
		ctx,
		s.Add,
		func(_ context.Context, key []byte, opts *pebble.WriteOptions) error {
			return s.DB.Delete(key, opts)
		},
		keys...,
	)
}

func (s *Store) keyDelete(
	ctx context.Context,
	add func() func(),
	del func(context.Context, []byte, *pebble.WriteOptions) error,
	keys ...string,
) (err error) {
	defer he(&err)
	defer add()()
	defer catchErr(&err, pebble.ErrClosed)

	for _, key := range keys {
		withMarshalKey(func(key []byte) {
			err = del(ctx, key, nil)
		}, Kv, key)
		ce(err,
			e5.Info("key %s", key),
		)
	}

	return nil
}

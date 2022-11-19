// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storestacked

import (
	"context"

	"github.com/reusee/e5"
	"github.com/reusee/june/key"
	"github.com/reusee/june/store"
	"github.com/reusee/sb"
)

var _ store.Store = new(Store)

func (s *Store) Exists(ctx context.Context, key Key) (bool, error) {
	select {
	case <-ctx.Done():
		return false, ErrClosed
	default:
	}

	switch s.ReadPolicy {

	case ReadThrough, ReadThroughCaching:
		if ok, err := s.Backing.Exists(ctx, key); err != nil {
			return false, err
		} else if ok {
			return true, nil
		}
		return s.Upstream.Exists(ctx, key)

	case ReadAround:
		return s.Upstream.Exists(ctx, key)

	}
	panic("bad policy")
}

func (s *Store) IterAllKeys(ctx context.Context, fn func(Key) error) (err error) {
	select {
	case <-ctx.Done():
		return ErrClosed
	default:
	}

	defer he(&err)

	switch s.ReadPolicy {

	case ReadThrough, ReadThroughCaching:
		// slow, but works
		backingKeys := make(map[Key]struct{})
		ce(s.Backing.IterAllKeys(ctx, func(key Key) error {
			backingKeys[key] = struct{}{}
			return nil
		}))
		isBreak := false
		ce(s.Upstream.IterAllKeys(ctx, func(key Key) (err error) {
			defer he(&err)
			delete(backingKeys, key)
			ce(
				fn(key),
				e5.WrapFunc(func(err error) error {
					if is(err, Break) {
						isBreak = true
					}
					return err
				}),
			)
			return nil
		}))
		if isBreak {
			return nil
		}
		for key := range backingKeys {
			ce(
				fn(key),
				e5.WrapFunc(func(err error) error {
					if is(err, Break) {
						return nil
					}
					return err
				}),
			)
		}
		return nil

	case ReadAround:
		return s.Upstream.IterAllKeys(ctx, fn)

	}
	panic("bad policy")
}

func (s *Store) IterKeys(ctx context.Context, ns key.Namespace, fn func(Key) error) (err error) {
	select {
	case <-ctx.Done():
		return ErrClosed
	default:
	}

	defer he(&err)

	switch s.ReadPolicy {

	case ReadThrough, ReadThroughCaching:
		// slow, but works
		backingKeys := make(map[Key]struct{})
		ce(s.Backing.IterKeys(ctx, ns, func(key Key) error {
			backingKeys[key] = struct{}{}
			return nil
		}))
		isBreak := false
		ce(s.Upstream.IterKeys(ctx, ns, func(key Key) (err error) {
			defer he(&err)
			delete(backingKeys, key)
			ce(
				fn(key),
				e5.WrapFunc(func(err error) error {
					if is(err, Break) {
						isBreak = true
					}
					return err
				}),
			)
			return nil
		}))
		if isBreak {
			return nil
		}
		for key := range backingKeys {
			ce(
				fn(key),
				e5.WrapFunc(func(err error) error {
					if is(err, Break) {
						return nil
					}
					return err
				}),
			)
		}
		return nil

	case ReadAround:
		return s.Upstream.IterKeys(ctx, ns, fn)

	}
	panic("bad policy")
}

func (s *Store) Read(ctx context.Context, key Key, fn func(sb.Stream) error) (err error) {
	select {
	case <-ctx.Done():
		return ErrClosed
	default:
	}

	defer he(&err)

	switch s.ReadPolicy {

	case ReadThrough:
		if err := s.Backing.Read(ctx, key, fn); err == nil {
			return nil
		} else if !is(err, ErrKeyNotFound) {
			return err
		}
		return s.Upstream.Read(ctx, key, fn)

	case ReadThroughCaching:
		if err := s.Backing.Read(ctx, key, fn); err == nil {
			return nil
		} else if !is(err, ErrKeyNotFound) {
			return err
		}
		return s.Upstream.Read(ctx, key, func(str sb.Stream) (err error) {
			defer he(&err)
			tokens, err := sb.TokensFromStream(str)
			ce(err)
			err = fn(tokens.Iter())
			ce(err)
			if _, err := s.Backing.Write(ctx, key.Namespace, tokens.Iter()); err != nil {
				if is(err, ErrIgnore) {
					err = nil
				} else {
					return err
				}
			}
			return nil
		})

	case ReadAround:
		return s.Upstream.Read(ctx, key, fn)

	}
	panic("bad policy")
}

func (s *Store) Write(
	ctx context.Context,
	ns key.Namespace,
	stream sb.Stream,
	options ...WriteOption,
) (res store.WriteResult, err error) {
	select {
	case <-ctx.Done():
		err = ErrClosed
		return
	default:
	}

	defer he(&err)

	switch s.WritePolicy {

	case WriteThrough:
		tokens, err := sb.TokensFromStream(stream)
		ce(err)
		res1, err := s.Backing.Write(ctx, ns, tokens.Iter(), options...)
		ignore := false
		if is(err, ErrIgnore) {
			ignore = true
			err = nil
		}
		ce(err)
		res2, err := s.Upstream.Write(ctx, ns, tokens.Iter(), options...)
		ce(err)
		if !ignore {
			if res1.Key != res2.Key {
				return res, we.With(e5.With(res1.Key))(ErrKeyNotMatch)
			}
		}
		return store.WriteResult{
			Key:          res2.Key,
			Written:      res1.Written || res2.Written,
			BytesWritten: res1.BytesWritten + res2.BytesWritten,
		}, nil

	case WriteAround:
		return s.Upstream.Write(ctx, ns, stream, options...)

	}
	panic("bad policy")
}

func (s *Store) Delete(ctx context.Context, keys []Key) (err error) {
	select {
	case <-ctx.Done():
		return ErrClosed
	default:
	}

	defer he(&err)

	switch s.WritePolicy {

	case WriteThrough:
		ce(s.Backing.Delete(ctx, keys))
		ce(s.Upstream.Delete(ctx, keys))
		return nil

	case WriteAround:
		ce(s.Upstream.Delete(ctx, keys))
		return nil

	}
	panic("bad policy")
}

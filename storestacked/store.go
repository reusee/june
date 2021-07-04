// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storestacked

import (
	"github.com/reusee/e4"
	"github.com/reusee/ling/v2/key"
	"github.com/reusee/ling/v2/store"
	"github.com/reusee/sb"
)

var _ store.Store = new(Store)

func (s *Store) Close() error {
	s.closeOnce.Do(func() {
		close(s.closed)
	})
	return nil
}

func (s *Store) Exists(key Key) (bool, error) {
	select {
	case <-s.closed:
		return false, ErrClosed
	default:
	}

	switch s.ReadPolicy {

	case ReadThrough, ReadThroughCaching:
		if ok, err := s.Backing.Exists(key); err != nil {
			return false, err
		} else if ok {
			return true, nil
		}
		return s.Upstream.Exists(key)

	case ReadAround:
		return s.Upstream.Exists(key)

	}
	panic("bad policy")
}

func (s *Store) IterAllKeys(fn func(Key) error) (err error) {
	select {
	case <-s.closed:
		return ErrClosed
	default:
	}

	defer he(&err)

	switch s.ReadPolicy {

	case ReadThrough, ReadThroughCaching:
		// slow, but works
		backingKeys := make(map[Key]struct{})
		ce(s.Backing.IterAllKeys(func(key Key) error {
			backingKeys[key] = struct{}{}
			return nil
		}))
		isBreak := false
		ce(s.Upstream.IterAllKeys(func(key Key) (err error) {
			defer he(&err)
			delete(backingKeys, key)
			ce(
				fn(key),
				func(err error) error {
					if is(err, Break) {
						isBreak = true
					}
					return err
				},
			)
			return nil
		}))
		if isBreak {
			return nil
		}
		for key := range backingKeys {
			ce(
				fn(key),
				func(err error) error {
					if is(err, Break) {
						return nil
					}
					return err
				},
			)
		}
		return nil

	case ReadAround:
		return s.Upstream.IterAllKeys(fn)

	}
	panic("bad policy")
}

func (s *Store) IterKeys(ns key.Namespace, fn func(Key) error) (err error) {
	select {
	case <-s.closed:
		return ErrClosed
	default:
	}

	defer he(&err)

	switch s.ReadPolicy {

	case ReadThrough, ReadThroughCaching:
		// slow, but works
		backingKeys := make(map[Key]struct{})
		ce(s.Backing.IterKeys(ns, func(key Key) error {
			backingKeys[key] = struct{}{}
			return nil
		}))
		isBreak := false
		ce(s.Upstream.IterKeys(ns, func(key Key) (err error) {
			defer he(&err)
			delete(backingKeys, key)
			ce(
				fn(key),
				func(err error) error {
					if is(err, Break) {
						isBreak = true
					}
					return err
				},
			)
			return nil
		}))
		if isBreak {
			return nil
		}
		for key := range backingKeys {
			ce(
				fn(key),
				func(err error) error {
					if is(err, Break) {
						return nil
					}
					return err
				},
			)
		}
		return nil

	case ReadAround:
		return s.Upstream.IterKeys(ns, fn)

	}
	panic("bad policy")
}

func (s *Store) Read(key Key, fn func(sb.Stream) error) (err error) {
	select {
	case <-s.closed:
		return ErrClosed
	default:
	}

	defer he(&err)

	switch s.ReadPolicy {

	case ReadThrough:
		if err := s.Backing.Read(key, fn); err == nil {
			return nil
		} else if !is(err, ErrKeyNotFound) {
			return err
		}
		return s.Upstream.Read(key, fn)

	case ReadThroughCaching:
		if err := s.Backing.Read(key, fn); err == nil {
			return nil
		} else if !is(err, ErrKeyNotFound) {
			return err
		}
		return s.Upstream.Read(key, func(str sb.Stream) (err error) {
			defer he(&err)
			tokens, err := sb.TokensFromStream(str)
			ce(err)
			err = fn(tokens.Iter())
			ce(err)
			if _, err := s.Backing.Write(key.Namespace, tokens.Iter()); err != nil {
				if is(err, ErrIgnore) {
					err = nil
				} else {
					return err
				}
			}
			return nil
		})

	case ReadAround:
		return s.Upstream.Read(key, fn)

	}
	panic("bad policy")
}

func (s *Store) Write(
	ns key.Namespace,
	stream sb.Stream,
	options ...WriteOption,
) (res store.WriteResult, err error) {
	select {
	case <-s.closed:
		err = ErrClosed
		return
	default:
	}

	defer he(&err)

	switch s.WritePolicy {

	case WriteThrough:
		tokens, err := sb.TokensFromStream(stream)
		ce(err)
		res1, err := s.Backing.Write(ns, tokens.Iter(), options...)
		ignore := false
		if is(err, ErrIgnore) {
			ignore = true
			err = nil
		}
		ce(err)
		res2, err := s.Upstream.Write(ns, tokens.Iter(), options...)
		ce(err)
		if !ignore {
			if res1.Key != res2.Key {
				return res, we(ErrKeyNotMatch, e4.With(res1.Key))
			}
		}
		return store.WriteResult{
			Key:          res2.Key,
			Written:      res1.Written || res2.Written,
			BytesWritten: res1.BytesWritten + res2.BytesWritten,
		}, nil

	case WriteAround:
		return s.Upstream.Write(ns, stream, options...)

	}
	panic("bad policy")
}

func (s *Store) Delete(keys []Key) (err error) {
	select {
	case <-s.closed:
		return ErrClosed
	default:
	}

	defer he(&err)

	switch s.WritePolicy {

	case WriteThrough:
		ce(s.Backing.Delete(keys))
		ce(s.Upstream.Delete(keys))
		return nil

	case WriteAround:
		ce(s.Upstream.Delete(keys))
		return nil

	}
	panic("bad policy")
}

func (s *Store) Sync() (err error) {
	select {
	case <-s.closed:
		return ErrClosed
	default:
	}
	defer he(&err)
	ce(s.Backing.Sync())
	ce(s.Upstream.Sync())
	return nil
}

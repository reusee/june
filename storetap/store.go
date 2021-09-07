// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storetap

import (
	"fmt"
	"sync/atomic"

	"github.com/reusee/june/key"
	"github.com/reusee/june/store"
	"github.com/reusee/sb"
)

type Store struct {
	name     string
	upstream store.Store
	funcs    Funcs
}

type Funcs struct {
	ID func(store.ID, error)

	Write func(
		key.Namespace,
		sb.Proc,
		[]store.WriteOption,
		store.WriteResult,
		error,
	)

	Read func(
		key.Key,
		func(sb.Proc) error,
		error,
	)

	Exists func(
		key.Key,
		bool,
		error,
	)

	IterKeys func(
		key.Namespace,
		func(key.Key) error,
		error,
	)

	IterAllKeys func(
		fn func(key.Key) error,
		err error,
	)

	Delete func(
		[]key.Key,
		error,
	)
}

type New func(upstream store.Store, funcs Funcs) *Store

func (_ Def) New() New {
	return func(upstream store.Store, funcs Funcs) *Store {
		return &Store{
			name: fmt.Sprintf("tap%d(%s)",
				atomic.AddInt64(&serial, 1),
				upstream.Name(),
			),
			upstream: upstream,
			funcs:    funcs,
		}
	}
}

var serial int64

var _ store.Store = new(Store)

func (s *Store) Name() string {
	return s.name
}

func (s *Store) ID() (id store.ID, err error) {
	defer func() {
		if s.funcs.ID != nil {
			s.funcs.ID(id, err)
		}
	}()
	return s.upstream.ID()
}

func (s *Store) Write(
	ns key.Namespace,
	proc sb.Proc,
	options ...store.WriteOption,
) (
	res store.WriteResult,
	err error,
) {
	defer func() {
		if s.funcs.Write != nil {
			s.funcs.Write(ns, proc, options, res, err)
		}
	}()
	return s.upstream.Write(ns, proc, options...)
}

func (s *Store) Read(
	key key.Key,
	fn func(sb.Proc) error,
) (err error) {
	defer func() {
		if s.funcs.Read != nil {
			s.funcs.Read(key, fn, err)
		}
	}()
	return s.upstream.Read(key, fn)
}

func (s *Store) Exists(
	key key.Key,
) (exists bool, err error) {
	defer func() {
		if s.funcs.Exists != nil {
			s.funcs.Exists(key, exists, err)
		}
	}()
	return s.upstream.Exists(key)
}

func (s *Store) IterKeys(
	ns key.Namespace,
	fn func(key.Key) error,
) (err error) {
	defer func() {
		if s.funcs.IterKeys != nil {
			s.funcs.IterKeys(ns, fn, err)
		}
	}()
	return s.upstream.IterKeys(ns, fn)
}

func (s *Store) IterAllKeys(
	fn func(key.Key) error,
) (err error) {
	defer func() {
		if s.funcs.IterAllKeys != nil {
			s.funcs.IterAllKeys(fn, err)
		}
	}()
	return s.upstream.IterAllKeys(fn)
}

func (s *Store) Delete(
	keys []key.Key,
) (err error) {
	defer func() {
		if s.funcs.Delete != nil {
			s.funcs.Delete(keys, err)
		}
	}()
	return s.upstream.Delete(keys)
}

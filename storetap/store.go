// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storetap

import (
	"context"
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
		sb.Stream,
		[]store.WriteOption,
		store.WriteResult,
		error,
	)

	Read func(
		key.Key,
		func(sb.Stream) error,
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

func (s *Store) ID(ctx context.Context) (id store.ID, err error) {
	defer func() {
		if s.funcs.ID != nil {
			s.funcs.ID(id, err)
		}
	}()
	return s.upstream.ID(ctx)
}

func (s *Store) Write(
	ctx context.Context,
	ns key.Namespace,
	stream sb.Stream,
	options ...store.WriteOption,
) (
	res store.WriteResult,
	err error,
) {
	defer func() {
		if s.funcs.Write != nil {
			s.funcs.Write(ns, stream, options, res, err)
		}
	}()
	return s.upstream.Write(ctx, ns, stream, options...)
}

func (s *Store) Read(
	ctx context.Context,
	key key.Key,
	fn func(sb.Stream) error,
) (err error) {
	defer func() {
		if s.funcs.Read != nil {
			s.funcs.Read(key, fn, err)
		}
	}()
	return s.upstream.Read(ctx, key, fn)
}

func (s *Store) Exists(
	ctx context.Context,
	key key.Key,
) (exists bool, err error) {
	defer func() {
		if s.funcs.Exists != nil {
			s.funcs.Exists(key, exists, err)
		}
	}()
	return s.upstream.Exists(ctx, key)
}

func (s *Store) IterKeys(
	ctx context.Context,
	ns key.Namespace,
	fn func(key.Key) error,
) (err error) {
	defer func() {
		if s.funcs.IterKeys != nil {
			s.funcs.IterKeys(ns, fn, err)
		}
	}()
	return s.upstream.IterKeys(ctx, ns, fn)
}

func (s *Store) IterAllKeys(
	ctx context.Context,
	fn func(key.Key) error,
) (err error) {
	defer func() {
		if s.funcs.IterAllKeys != nil {
			s.funcs.IterAllKeys(fn, err)
		}
	}()
	return s.upstream.IterAllKeys(ctx, fn)
}

func (s *Store) Delete(
	ctx context.Context,
	keys []key.Key,
) (err error) {
	defer func() {
		if s.funcs.Delete != nil {
			s.funcs.Delete(keys, err)
		}
	}()
	return s.upstream.Delete(ctx, keys)
}

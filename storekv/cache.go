// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storekv

import (
	"context"

	"github.com/reusee/e5"
	"github.com/reusee/june/store"
	"github.com/reusee/sb"
)

var _ Cache = new(Store)

func (s *Store) CacheGet(ctx context.Context, key Key, fn func(sb.Stream) error) (err error) {
	defer he(&err)
	if err := s.Read(ctx, key, func(s sb.Stream) error {
		return fn(s)
	}); is(err, ErrKeyNotFound) {
		return we.With(e5.With(key))(err)
	} else {
		ce(err)
	}
	return nil
}

func (s *Store) CachePut(ctx context.Context, key Key, tokes sb.Tokens, options ...store.CachePutOption) error {
	// ignore
	return nil
}

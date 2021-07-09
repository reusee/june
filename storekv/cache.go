// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storekv

import (
	"github.com/reusee/e4"
	"github.com/reusee/june/store"
	"github.com/reusee/sb"
)

var _ Cache = new(Store)

func (s *Store) CacheGet(key Key, fn func(sb.Stream) error) (err error) {
	defer he(&err)
	if err := s.Read(key, func(s sb.Stream) error {
		return fn(s)
	}); is(err, ErrKeyNotFound) {
		return we(err, e4.With(key))
	} else {
		ce(err)
	}
	return nil
}

func (s *Store) CachePut(key Key, tokes sb.Tokens, options ...store.CachePutOption) error {
	// ignore
	return nil
}

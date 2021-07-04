// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storemem

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/google/btree"
)

type Store struct {
	name   string
	index  *btree.BTree
	values sync.Map
	sync.RWMutex
	closed    chan struct{}
	closeOnce sync.Once
}

type New func() *Store

func (_ Def) New() New {
	return func() *Store {
		return &Store{
			name:   fmt.Sprintf("mem%d", atomic.AddInt64(&serial, 1)),
			index:  btree.New(2),
			closed: make(chan struct{}),
		}
	}
}

var serial int64

func (s *Store) Name() string {
	return s.name
}

func (s *Store) StoreID() string {
	return s.name
}

func (s *Store) Close() error {
	s.closeOnce.Do(func() {
		close(s.closed)
	})
	return nil
}

// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storemem

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/google/btree"
	"github.com/reusee/pr"
)

type Store struct {
	*pr.WaitTree
	name   string
	index  *btree.BTree
	values sync.Map
	sync.RWMutex
}

type New func() *Store

func (_ Def) New(
	parentWt *pr.WaitTree,
) New {
	return func() *Store {
		return &Store{
			WaitTree: parentWt,
			name:     fmt.Sprintf("mem%d", atomic.AddInt64(&serial, 1)),
			index:    btree.New(2),
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

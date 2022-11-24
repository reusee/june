// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storemem

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/google/btree"
	"github.com/reusee/pr2"
	"github.com/reusee/sb"
)

type Store struct {
	wg     *pr2.WaitGroup
	name   string
	index  *btree.BTreeG[sb.Tokens]
	values sync.Map
	sync.RWMutex
}

type New func(ctx context.Context) *Store

func (Def) New() New {
	return func(
		ctx context.Context,
	) *Store {
		wg := pr2.NewWaitGroup(ctx)
		return &Store{
			wg:   wg,
			name: fmt.Sprintf("mem%d", atomic.AddInt64(&serial, 1)),
			index: btree.NewG(2, func(a, b sb.Tokens) bool {
				return sb.MustCompare(a.Iter(), b.Iter()) < 0
			}),
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

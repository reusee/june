// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storestacked

import (
	"fmt"
	"sync/atomic"

	"github.com/reusee/june/store"
	"github.com/reusee/pr"
)

type Store struct {
	*pr.WaitTree
	name        string
	Upstream    store.Store
	Backing     store.Store
	id          StoreID
	ReadPolicy  ReadPolicy
	WritePolicy WritePolicy
}

type WritePolicy uint8

const (
	WriteThrough WritePolicy = iota
	WriteAround
)

type ReadPolicy uint8

const (
	ReadThrough ReadPolicy = iota
	ReadThroughCaching
	ReadAround
)

type New func(
	store.Store,
	store.Store,
	ReadPolicy,
	WritePolicy,
) (*Store, error)

func (_ Def) New(
	parentWt *pr.WaitTree,
) New {
	return func(
		upstream store.Store,
		backing store.Store,
		readPolicy ReadPolicy,
		writePolicy WritePolicy,
	) (_ *Store, err error) {
		defer he(&err)

		id1, err := upstream.ID()
		ce(err)
		id2, err := backing.ID()
		ce(err)

		wt := pr.NewWaitTree(parentWt)

		return &Store{
			WaitTree: wt,
			name: fmt.Sprintf("stacked%d(%s, %s)",
				atomic.AddInt64(&serial, 1),
				upstream.Name(),
				backing.Name(),
			),
			id:          "(stacked(" + id1 + "," + id2 + "))",
			Upstream:    upstream,
			Backing:     backing,
			ReadPolicy:  readPolicy,
			WritePolicy: writePolicy,
		}, nil
	}
}

var serial int64

func (s *Store) Name() string {
	return s.name
}

func (s *Store) ID() (StoreID, error) {
	return s.id, nil
}

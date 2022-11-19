// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storemonotree

import (
	"context"
	"testing"

	"github.com/reusee/dscope"
	"github.com/reusee/e5"
	"github.com/reusee/june/store"
	"github.com/reusee/june/storekv"
	"github.com/reusee/june/storemem"
)

func TestStore(
	t *testing.T,
	testStore store.TestStore,
	scope dscope.Scope,
) {
	t.Skip() //TODO

	defer he(nil, e5.TestingFatal(t))

	ctx := context.Background()
	with := func(fn func(store.Store), defs ...any) {
		scope.Fork(defs...).Call(func(
			newMem storemem.New,
			newKV storekv.New,
			newTree New,
		) {
			upstream, err := newKV(newMem(), "foo")
			ce(err)
			tree, err := newTree(ctx, upstream)
			ce(err)
			fn(tree)
		})
	}
	testStore(ctx, with, t)
}

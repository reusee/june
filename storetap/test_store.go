// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storetap

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
	defer e5.Handle(nil, e5.TestingFatal(t))
	ctx := context.Background()

	with := func(
		fn func(store.Store),
		defs ...any,
	) {
		scope.Fork(defs...).Call(func(
			newMem storemem.New,
			newKV storekv.New,
			newStore New,
		) {
			upstream, err := newKV(newMem(), "foo")
			ce(err)
			store := newStore(upstream, Funcs{})
			fn(store)
		})
	}

	testStore(ctx, with, t)

}

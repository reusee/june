// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storetap

import (
	"testing"

	"github.com/reusee/dscope"
	"github.com/reusee/e5"
	"github.com/reusee/june/store"
	"github.com/reusee/june/storekv"
	"github.com/reusee/june/storemem"
	"github.com/reusee/pr2"
)

func TestStore(
	t *testing.T,
	testStore store.TestStore,
	scope dscope.Scope,
	wg *pr2.WaitGroup,
) {
	defer e5.Handle(nil, e5.TestingFatal(t))

	with := func(
		fn func(store.Store),
		defs ...any,
	) {
		scope.Fork(defs...).Call(func(
			newMem storemem.New,
			newKV storekv.New,
			newStore New,
		) {
			upstream, err := newKV(wg, newMem(wg), "foo")
			ce(err)
			store := newStore(upstream, Funcs{})
			fn(store)
		})
	}

	testStore(wg, with, t)

}

// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storetap

import (
	"testing"

	"github.com/reusee/dscope"
	"github.com/reusee/e4"
	"github.com/reusee/june/store"
	"github.com/reusee/june/storekv"
	"github.com/reusee/june/storemem"
)

func TestStore(
	t *testing.T,
	testStore store.TestStore,
	scope dscope.Scope,
) {
	defer e4.Handle(nil, e4.TestingFatal(t))

	with := func(
		fn func(store.Store),
		defs ...any,
	) {
		scope.Sub(defs...).Call(func(
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

	testStore(with, t)

}

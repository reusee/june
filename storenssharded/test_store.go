// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storenssharded

import (
	"testing"

	"github.com/reusee/dscope"
	"github.com/reusee/e5"
	"github.com/reusee/june/key"
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
	defer he(nil, e5.TestingFatal(t))

	with := func(fn func(store.Store), defs ...any) {
		scope.Fork(defs...).Call(func(
			newMem storemem.New,
			newKV storekv.New,
			newStore New,
		) {
			s1, err := newKV(wg, newMem(wg), "foo")
			ce(err)
			s2, err := newKV(wg, newMem(wg), "foo")
			ce(err)
			s, err := newStore(
				map[key.Namespace]store.Store{
					{'f', 'o', 'o'}: s1,
					{'b', 'a', 'r'}: s1,
				},
				s2,
			)
			ce(err)
			fn(s)
		})
	}
	testStore(wg, with, t)
}

// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storestacked

import (
	"context"
	"fmt"
	"testing"

	"github.com/reusee/dscope"
	"github.com/reusee/june/store"
	"github.com/reusee/june/storekv"
	"github.com/reusee/june/storemem"
)

func TestStore(
	t *testing.T,
	testStore store.TestStore,
	scope dscope.Scope,
) {
	for _, readPolicy := range []ReadPolicy{
		ReadThrough,
		ReadThroughCaching,
		ReadAround,
	} {
		for _, writePolicy := range []WritePolicy{
			WriteThrough,
			WriteAround,
		} {

			t.Run(fmt.Sprintf("%v / %v", readPolicy, writePolicy), func(t *testing.T) {
				ctx := context.Background()
				with := func(fn func(store.Store), defs ...any) {
					scope.Fork(defs...).Call(func(
						newMem storemem.New,
						newKV storekv.New,
						newStore New,
					) {
						backing, err := newKV(newMem(), "foo")
						if err != nil {
							t.Fatal(err)
						}
						upstream, err := newKV(newMem(), "foo")
						if err != nil {
							t.Fatal(err)
						}
						store, err := newStore(ctx, upstream, backing, readPolicy, writePolicy)
						ce(err)
						fn(store)
					})
				}
				testStore(ctx, with, t)
			})

		}
	}
}

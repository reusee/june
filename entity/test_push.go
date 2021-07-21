// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package entity

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/reusee/e4"
	"github.com/reusee/june/storekv"
	"github.com/reusee/june/storemem"
	"github.com/reusee/pr"
)

func TestPush(
	t *testing.T,
	wt *pr.WaitTree,
	newMem storemem.New,
	newKV storekv.New,
	scope Scope,
	rootCtx context.Context,
) {
	defer he(nil, e4.TestingFatal(t))

	mem1 := newMem(wt)
	store1, err := newKV(mem1, "foo")
	ce(err)
	scope.Sub(func() Store {
		return store1
	}).Call(func(
		saveEntity SaveEntity,
		push Push,
		indexManager IndexManager,
	) {

		summary, err := saveEntity(42)
		ce(err)
		key1 := summary.Key
		type Keys []Key
		summary, err = saveEntity(Keys{
			key1, key1,
		})
		ce(err)
		key2 := summary.Key
		summary, err = saveEntity(Keys{
			key1, key1, key2,
		})
		ce(err)
		key := summary.Key

		{
			mem2 := newMem(wt)
			store2, err := newKV(mem2, "foo")
			ce(err)

			var numCheck int64
			var numSave int64
			err = push(
				store2, indexManager,
				[]Key{key},
				TapPushCheckSummary(func(_ Key) {
					atomic.AddInt64(&numCheck, 1)
				}),
				TapPushSave(func(_ Key, _ *Summary) {
					atomic.AddInt64(&numSave, 1)
				}),
			)
			ce(err)
			if numCheck != 3 {
				t.Fatal()
			}
			if numSave != 3 {
				t.Fatal()
			}

			scope.Sub(func() Store {
				return store2
			}).Call(func(
				checkRef CheckRef,
				store Store,
			) {

				n := 0
				ce(store.IterAllKeys(func(key Key) error {
					n++
					return nil
				}))
				if n != 6 {
					t.Fatalf("got %d", n)
				}

				ce(checkRef())
			})

		}

		// keys not specified
		{
			mem2 := newMem(wt)
			store2, err := newKV(mem2, "foo")
			ce(err)

			var numCheck int64
			var numSave int64
			err = push(
				store2, indexManager,
				nil,
				TapPushCheckSummary(func(_ Key) {
					atomic.AddInt64(&numCheck, 1)
				}),
				TapPushSave(func(_ Key, _ *Summary) {
					atomic.AddInt64(&numSave, 1)
				}),
			)
			ce(err)
			if numCheck != 3 {
				t.Fatal()
			}
			if numSave != 3 {
				t.Fatal()
			}

			scope.Sub(func() Store {
				return store2
			}).Call(func(
				checkRef CheckRef,
				store Store,
			) {

				n := 0
				ce(store.IterAllKeys(func(key Key) error {
					n++
					return nil
				}))
				if n != 6 {
					t.Fatalf("got %d", n)
				}

				ce(checkRef())
			})

		}

		// index manager not specified
		{
			mem2 := newMem(wt)
			store2, err := newKV(mem2, "foo")
			ce(err)

			var numCheck int64
			var numSave int64
			err = push(
				store2, nil,
				nil,
				TapPushCheckSummary(func(_ Key) {
					atomic.AddInt64(&numCheck, 1)
				}),
				TapPushSave(func(_ Key, _ *Summary) {
					atomic.AddInt64(&numSave, 1)
				}),
			)
			ce(err)
			if numCheck != 3 {
				t.Fatal()
			}
			if numSave != 3 {
				t.Fatal()
			}

			scope.Sub(func() Store {
				return store2
			}).Call(func(
				checkRef CheckRef,
				store Store,
			) {

				n := 0
				ce(store.IterAllKeys(func(key Key) error {
					n++
					return nil
				}))
				if n != 6 {
					t.Fatalf("got %d", n)
				}

				ce(checkRef())
			})

		}

	})

}

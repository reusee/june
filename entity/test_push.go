// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package entity

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/reusee/e5"
	"github.com/reusee/june/storekv"
	"github.com/reusee/june/storemem"
)

func TestPush(
	t *testing.T,
	newMem storemem.New,
	newKV storekv.New,
	scope Scope,
) {
	defer he(nil, e5.TestingFatal(t))
	ctx := context.Background()

	mem1 := newMem()
	store1, err := newKV(mem1, "foo")
	ce(err)
	scope.Fork(func() Store {
		return store1
	}).Call(func(
		saveEntity SaveEntity,
		push Push,
		indexManager IndexManager,
	) {

		summary, err := saveEntity(ctx, 42)
		ce(err)
		key1 := summary.Key
		type Keys []Key
		summary, err = saveEntity(ctx, Keys{
			key1, key1,
		})
		ce(err)
		key2 := summary.Key
		summary, err = saveEntity(ctx, Keys{
			key1, key1, key2,
		})
		ce(err)
		key := summary.Key

		{
			mem2 := newMem()
			store2, err := newKV(mem2, "foo")
			ce(err)

			var numCheck int64
			var numSave int64
			err = push(
				ctx,
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
				t.Fatalf("got %d", numCheck)
			}
			if numSave != 3 {
				t.Fatal()
			}

			scope.Fork(func() Store {
				return store2
			}).Call(func(
				checkRef CheckRef,
				store Store,
			) {

				n := 0
				ce(store.IterAllKeys(ctx, func(_ Key) error {
					n++
					return nil
				}))
				if n != 6 {
					t.Fatalf("got %d", n)
				}

				ce(checkRef(ctx))
			})

		}

		// keys not specified
		{
			mem2 := newMem()
			store2, err := newKV(mem2, "foo")
			ce(err)

			var numCheck int64
			var numSave int64
			err = push(
				ctx,
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
				t.Fatalf("got %d\n", numCheck)
			}
			if numSave != 3 {
				t.Fatal()
			}

			scope.Fork(func() Store {
				return store2
			}).Call(func(
				checkRef CheckRef,
				store Store,
			) {

				n := 0
				ce(store.IterAllKeys(ctx, func(_ Key) error {
					n++
					return nil
				}))
				if n != 6 {
					t.Fatalf("got %d", n)
				}

				ce(checkRef(ctx))
			})

		}

		// index manager not specified
		{
			mem2 := newMem()
			store2, err := newKV(mem2, "foo")
			ce(err)

			var numCheck int64
			var numSave int64
			err = push(
				ctx,
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
				t.Fatalf("got %d", numCheck)
			}
			if numSave != 3 {
				t.Fatal()
			}

			scope.Fork(func() Store {
				return store2
			}).Call(func(
				checkRef CheckRef,
				store Store,
			) {

				n := 0
				ce(store.IterAllKeys(ctx, func(_ Key) error {
					n++
					return nil
				}))
				if n != 6 {
					t.Fatalf("got %d", n)
				}

				ce(checkRef(ctx))
			})

		}

		// ignore
		{
			mem2 := newMem()
			store2, err := newKV(mem2, "foo")
			ce(err)

			var numCheck int64
			var numSave int64
			err = push(
				ctx,
				store2, indexManager,
				[]Key{key},
				TapPushCheckSummary(func(_ Key) {
					atomic.AddInt64(&numCheck, 1)
				}),
				TapPushSave(func(_ Key, _ *Summary) {
					atomic.AddInt64(&numSave, 1)
				}),
				IgnoreSummary(func(s Summary) bool {
					return false
				}),
			)
			ce(err)
			if numCheck != 3 {
				t.Fatalf("got %d", numCheck)
			}
			if numSave != 3 {
				t.Fatal()
			}

			scope.Fork(func() Store {
				return store2
			}).Call(func(
				checkRef CheckRef,
				store Store,
			) {

				n := 0
				ce(store.IterAllKeys(ctx, func(_ Key) error {
					n++
					return nil
				}))
				if n != 6 {
					t.Fatalf("got %d", n)
				}

				ce(checkRef(ctx))
			})

		}

	})

}

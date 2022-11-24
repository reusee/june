// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package entity

import (
	"sync/atomic"
	"testing"

	"github.com/reusee/e5"
	"github.com/reusee/june/storekv"
	"github.com/reusee/june/storemem"
	"github.com/reusee/pr2"
)

func TestPush(
	t *testing.T,
	newMem storemem.New,
	newKV storekv.New,
	scope Scope,
	wg *pr2.WaitGroup,
) {
	defer he(nil, e5.TestingFatal(t))

	mem1 := newMem(wg)
	store1, err := newKV(wg, mem1, "foo")
	ce(err)
	scope.Fork(func() Store {
		return store1
	}).Call(func(
		saveEntity SaveEntity,
		push Push,
		indexManager IndexManager,
	) {

		summary, err := saveEntity(wg, 42)
		ce(err)
		key1 := summary.Key
		type Keys []Key
		summary, err = saveEntity(wg, Keys{
			key1, key1,
		})
		ce(err)
		key2 := summary.Key
		summary, err = saveEntity(wg, Keys{
			key1, key1, key2,
		})
		ce(err)
		key := summary.Key

		{
			mem2 := newMem(wg)
			store2, err := newKV(wg, mem2, "foo")
			ce(err)

			var numCheck int64
			var numSave int64
			err = push(
				wg,
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
				ce(store.IterAllKeys(func(_ Key) error {
					n++
					return nil
				}))
				if n != 6 {
					t.Fatalf("got %d", n)
				}

				ce(checkRef(wg))
			})

		}

		// keys not specified
		{
			mem2 := newMem(wg)
			store2, err := newKV(wg, mem2, "foo")
			ce(err)

			var numCheck int64
			var numSave int64
			err = push(
				wg,
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
				ce(store.IterAllKeys(func(_ Key) error {
					n++
					return nil
				}))
				if n != 6 {
					t.Fatalf("got %d", n)
				}

				ce(checkRef(wg))
			})

		}

		// index manager not specified
		{
			mem2 := newMem(wg)
			store2, err := newKV(wg, mem2, "foo")
			ce(err)

			var numCheck int64
			var numSave int64
			err = push(
				wg,
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
				ce(store.IterAllKeys(func(_ Key) error {
					n++
					return nil
				}))
				if n != 6 {
					t.Fatalf("got %d", n)
				}

				ce(checkRef(wg))
			})

		}

		// ignore
		{
			mem2 := newMem(wg)
			store2, err := newKV(wg, mem2, "foo")
			ce(err)

			var numCheck int64
			var numSave int64
			err = push(
				wg,
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
				ce(store.IterAllKeys(func(_ Key) error {
					n++
					return nil
				}))
				if n != 6 {
					t.Fatalf("got %d", n)
				}

				ce(checkRef(wg))
			})

		}

	})

}

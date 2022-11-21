// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storekv

import (
	"io"
	"testing"

	"github.com/reusee/dscope"
	"github.com/reusee/e5"
	"github.com/reusee/june/store"
)

type TestKV func(
	t *testing.T,
	with func(
		fn func(kv KV, prefix string),
	),
)

func (_ Def) TestKV(
	scope dscope.Scope,
	testStore store.TestStore,
) TestKV {
	return func(
		t *testing.T,
		with func(
			kvFunc func(kv KV, prefix string),
		),
	) {
		defer he(nil, e5.TestingFatal(t))

		withStore := func(storeFunc func(store.Store), provides ...any) {
			with(func(kv KV, prefix string) {
				scope.Fork(provides...).Call(func(
					newStore New,
					codec Codec,
				) {
					store, err := newStore(
						kv, prefix,
						WithCodec(codec),
					)
					ce(err)
					storeFunc(store)
				})
			})
		}
		testStore(withStore, t)

		// cache
		withStore = func(storeFunc func(store.Store), provides ...any) {
			with(func(kv KV, prefix string) {
				scope.Fork(provides...).Call(func(
					newStore New,
					codec Codec,
					newMemCache store.NewMemCache,
				) {
					cache, err := newMemCache(1024, 8192)
					ce(err)
					store, err := newStore(
						kv, prefix,
						WithCodec(codec),
						WithCache(cache),
					)
					ce(err)
					storeFunc(store)
				})
			})
		}
		testStore(withStore, t)

		// offload
		withStore = func(storeFunc func(store.Store), provides ...any) {
			with(func(offloadKV KV, prefix string) {
				scope.Call(func(
					newStore New,
				) {
					offloadStore, err := newStore(offloadKV, "offload")
					ce(err)
					with(func(kv KV, prefix string) {
						scope.Fork(provides...).Call(func(
							newStore New,
							codec Codec,
						) {
							store, err := newStore(
								kv,
								prefix,
								WithCodec(codec),
								WithOffload(func(key Key, l int) store.Store {
									return offloadStore
								}),
							)
							ce(err)
							storeFunc(store)
						})
					})
				})
			})
		}
		testStore(withStore, t)

		// errors
		with(func(kv KV, prefix string) {
			key := "foo"
			err := kv.KeyGet(key, func(r io.Reader) error {
				_, err := io.Copy(io.Discard, r)
				return err
			})

			if !is(err, ErrKeyNotFound) {
				t.Fatalf("got %#v", err)
			}
			var errKey StringKey
			if !as(err, &errKey) {
				t.Fatal()
			}
			if string(errKey) != key {
				t.Fatal()
			}

		})

	}
}

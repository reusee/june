// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package store

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/reusee/e5"
	"github.com/reusee/june/codec"
	"github.com/reusee/june/key"
	"github.com/reusee/june/opts"
	"github.com/reusee/sb"
)

// test Store implementation
type TestStore func(
	withStore func(
		fn func(Store),
		provides ...any,
	),
	t *testing.T,
)

func (Def) TestStore(
	scrub Scrub,
) TestStore {

	return func(
		withStore func(
			fn func(Store),
			provides ...any,
		),
		t *testing.T,
	) {
		defer he(nil, e5.TestingFatal(t))

		ns := key.Namespace{'f', 'o', 'o'}

		fn := func(store Store) {
			var res WriteResult
			t.Run("write", func(t *testing.T) {
				var err error
				res, err = store.Write(ns, sb.Marshal(42))
				ce(err)
				if res.Key.Hash.String() != "151a3a0b4c88483512fc484d0badfedf80013ebb18df498bbee89ac5b69d7222" {
					t.Fatalf("got %x", res.Key)
				}
				if res.Key.Namespace != ns {
					t.Fatal()
				}
			})

			t.Run("read", func(t *testing.T) {
				defer he(nil, e5.TestingFatal(t))
				if err := store.Read(res.Key, func(s sb.Stream) (err error) {
					defer he(&err)
					var i int
					err = sb.Copy(s, sb.Unmarshal(&i))
					ce(err)
					if i != 42 {
						t.Fatalf("not 42")
					}
					return nil
				}); err != nil {
					t.Fatal(err)
				}
			})

			t.Run("read error", func(t *testing.T) {
				errFoo := fmt.Errorf("bad")
				err := store.Read(res.Key, func(s sb.Stream) error {
					return errFoo
				})
				if err == nil {
					t.Fatal()
				}
				if !is(err, ErrRead) {
					t.Fatal()
				}
				var errKey Key
				if !as(err, &errKey) {
					t.Fatal()
				}
				if errKey != res.Key {
					t.Fatal()
				}
				if !is(err, errFoo) {
					t.Fatal()
				}
			})

			t.Run("exists", func(t *testing.T) {
				defer he(nil, e5.TestingFatal(t))
				ok, err := store.Exists(res.Key)
				ce(err)
				if !ok {
					t.Fatalf("should exists")
				}
				ok, err = store.Exists(Key{Hash: Hash{1, 2, 3}})
				ce(err)
				if ok {
					t.Fatalf("should not exists")
				}
			})

			num := int64(rand.Intn(16) + 16)
			t.Run("more write", func(t *testing.T) {
				wg := new(sync.WaitGroup)
				wg.Add(int(num))
				errors := make(chan error, num)
				for i := int64(0); i < num; i++ {
					go func() {
						defer wg.Done()
						_, err := store.Write(ns, sb.Marshal(rand.Int63()))
						if err != nil {
							errors <- err
							return
						}
					}()
				}
				wg.Wait()
				select {
				case err := <-errors:
					t.Fatal(err)
				default:
				}
			})

			t.Run("iter", func(t *testing.T) {
				var n int64
				if err := store.IterKeys(ns, func(_ Key) error {
					atomic.AddInt64(&n, 1)
					return nil
				}); err != nil {
					t.Fatal(err)
				}
				if n != num+1 {
					t.Fatalf("got %d, expected %d\n", n, num+1)
				}
			})

			t.Run("iter all", func(t *testing.T) {
				var n int64
				if err := store.IterAllKeys(func(_ Key) error {
					atomic.AddInt64(&n, 1)
					return nil
				}); err != nil {
					t.Fatal(err)
				}
				if n != num+1 {
					t.Fatalf("got %d, expected %d\n", n, num+1)
				}
			})

			t.Run("iter break", func(t *testing.T) {
				var n int64
				if err := store.IterKeys(ns, func(_ Key) error {
					if atomic.AddInt64(&n, 1) == num/2 {
						return Break
					}
					return nil
				}); err != nil {
					t.Fatal(err)
				}
				if n != num/2 {
					t.Fatal()
				}
			})

			e := errors.New("foo")
			t.Run("iter err", func(t *testing.T) {
				var n int64
				if err := store.IterKeys(ns, func(_ Key) error {
					if atomic.AddInt64(&n, 1) == num/2 {
						return e
					}
					return nil
				}); !is(err, e) {
					t.Fatal()
				}
				if n != num/2 {
					t.Fatal()
				}
			})

			t.Run("scrub", func(t *testing.T) {
				var n int64
				if err := scrub(
					store,
					opts.TapKey(func(_ Key) {
						atomic.AddInt64(&n, 1)
					}),
					opts.TapBadKey(func(_ Key) {
						t.Fatal("bad store")
					}),
				); err != nil {
					t.Fatal(err)
				}
				if n != int64(num+1) {
					t.Fatalf("got %d\n", n)
				}
			})

			t.Run("namespace", func(t *testing.T) {
				defer he(nil, e5.TestingFatal(t))
				nsBar := key.Namespace{'b', 'a', 'r'}
				res, err := store.Write(nsBar, sb.Marshal(42))
				ce(err)
				ok, err := store.Exists(res.Key)
				ce(err)
				if !ok {
					t.Fatal()
				}
				var n int64
				if err := store.IterKeys(nsBar, func(_ Key) error {
					atomic.AddInt64(&n, 1)
					return nil
				}); err != nil {
					t.Fatal(err)
				}
				if n != 1 {
					t.Fatalf("got %d\n", n)
				}
				n = 0
				nss := make(map[key.Namespace]bool)
				if err := store.IterAllKeys(func(key Key) error {
					n++
					nss[key.Namespace] = true
					return nil
				}); err != nil {
					t.Fatal(err)
				}
				if n != num+2 {
					t.Fatalf("got %d", n)
				}
				if len(nss) != 2 {
					t.Fatalf("got %d", len(nss))
				}
				if !nss[ns] {
					t.Fatal()
				}
				if !nss[nsBar] {
					t.Fatal()
				}
			})

			t.Run("options", func(t *testing.T) {
				defer he(nil, e5.TestingFatal(t))
				withKeyOK := false
				withResultOK := false
				_, err := store.Write(
					ns, sb.Marshal(42),
					opts.TapKey(func(_ Key) {
						withKeyOK = true
					}),
					TapWriteResult(func(_ WriteResult) {
						withResultOK = true
					}),
				)
				ce(err)
				if !withKeyOK {
					t.Fatal()
				}
				if !withResultOK {
					t.Fatal()
				}
			})

			t.Run("delete", func(t *testing.T) {
				defer he(nil, e5.TestingFatal(t))
				res1, err := store.Write(
					ns, sb.Marshal(rand.Int63()),
				)
				ce(err)
				res2, err := store.Write(
					ns, sb.Marshal(rand.Int63()),
				)
				ce(err)
				ok, err := store.Exists(res1.Key)
				ce(err)
				if !ok {
					t.Fatal()
				}
				ok, err = store.Exists(res2.Key)
				ce(err)
				if !ok {
					t.Fatal()
				}
				err = store.Delete([]Key{res1.Key, res2.Key})
				ce(err)
				ok, err = store.Exists(res1.Key)
				ce(err)
				if ok {
					t.Fatal()
				}
				ok, err = store.Exists(res2.Key)
				ce(err)
				if ok {
					t.Fatal()
				}
			})

		}

		withStore(
			func(store Store) {
				fn(store)
			},
			func() codec.Codec {
				return codec.DefaultCodec
			},
		)

		// codec
		key := []byte("0123456789012345")
		newAEAD := func() (_ cipher.AEAD, err error) {
			defer he(&err)
			block, err := aes.NewCipher(key)
			ce(err)
			aead, err := cipher.NewGCM(block)
			ce(err)
			return aead, nil
		}
		withStore(
			fn,
			func() codec.Codec {
				return codec.NewAEADCodec(
					fmt.Sprintf("aead-%x", key),
					newAEAD,
				)
			},
		)

		// concurrent read and write
		withStore(
			func(
				store Store,
			) {
				_, err := store.Write(ns, sb.Marshal(42))
				ce(err)
				if err := store.IterAllKeys(func(_ Key) (err error) {
					done := make(chan struct{})
					go func() {
						defer func() {
							close(done)
						}()
						_, err = store.Write(ns, sb.Marshal(rand.Int63()))
					}()
					select {
					case <-done:
					case <-time.After(time.Second * 1):
						t.Fatal("dead lock")
					}
					return
				}); err != nil {
					t.Fatal(err)
				}
			},
			func() codec.Codec {
				return codec.DefaultCodec
			},
		)

	}
}

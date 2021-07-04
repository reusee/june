// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package tx

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/reusee/e4"
	"github.com/reusee/june/entity"
	"github.com/reusee/june/filebase"
	"github.com/reusee/june/index"
	"github.com/reusee/june/store"
	"github.com/reusee/june/storekv"
	"github.com/reusee/june/storepebble"
)

func TestPebbleTx(
	t *testing.T,
	newPeb storepebble.New,
	newKV storekv.New,
	scope Scope,
) {
	defer he(nil, e4.TestingFatal(t))

	dir := t.TempDir()
	peb, err := newPeb(nil, dir)
	ce(err)
	defer peb.Close()

	kv, err := newKV(peb, "foo")
	ce(err)
	defer kv.Close()

	scope.Sub(
		func() KVToStore {
			return func(kv storekv.KV) (store.Store, error) {
				return newKV(kv, "foo")
			}
		},
		UsePebbleTx,
		&peb,
	).Call(func(
		tx PebbleTx,
		scope Scope,
	) {

		// commit tx
		var key1 Key
		ce(tx(func(
			save entity.Save,
		) {
			summary, err := save(entity.NSEntity, 42)
			ce(err)
			key1 = summary.Key
		}))

		ce(tx(func(
			fetch entity.Fetch,
			selIndex index.SelectIndex,
		) {
			var i int
			ce(fetch(key1, &i))
			if i != 42 {
				t.Fatal()
			}

			ce(selIndex(
				index.MatchEntry(entity.IdxSummaryKey, key1),
				index.Count(&i),
			))
			if i != 1 {
				t.Fatal()
			}
		}))

		// error, no commit
		errFoo := fmt.Errorf("foo")
		var key2 Key
		err = tx(func(
			save entity.Save,
		) {
			summary, err := save(entity.NSEntity, 1)
			ce(err)
			key2 = summary.Key
			ce(errFoo)
		})
		if !errors.Is(err, errFoo) {
			t.Fatal()
		}

		ce(tx(func(
			fetch entity.Fetch,
			selIndex index.SelectIndex,
		) {
			var i int
			err := fetch(key2, &i)
			if !errors.Is(err, store.ErrKeyNotFound) {
				t.Fatal()
			}

			ce(selIndex(
				index.MatchEntry(entity.IdxSummaryKey, key2),
				index.Count(&i),
			))
			if i != 0 {
				t.Fatal()
			}
			ce(selIndex(
				index.MatchEntry(entity.IdxSummaryKey, key1),
				index.Count(&i),
			))
			if i != 1 {
				t.Fatal()
			}
		}))

		// tx inside tx, partial commit
		var key3, key4 Key
		err = tx(func(
			save entity.Save,
			store store.Store,
		) {
			ce(tx(func(
				save entity.Save,
			) {
				summary, err := save(entity.NSEntity, 99)
				ce(err)
				key3 = summary.Key
			}))

			// should see committed key
			ok, err := store.Exists(key3)
			ce(err)
			if !ok {
				t.Fatal()
			}

			summary, err := save(entity.NSEntity, 1)
			ce(err)
			key4 = summary.Key
			ce(errFoo)
		})
		if !errors.Is(err, errFoo) {
			t.Fatal()
		}

		ce(tx(func(
			fetch entity.Fetch,
			selIndex index.SelectIndex,
		) {
			var i int

			ce(fetch(key3, &i))
			if i != 99 {
				t.Fatal()
			}

			err := fetch(key4, &i)
			if !errors.Is(err, store.ErrKeyNotFound) {
				t.Fatal()
			}

			ce(selIndex(
				index.MatchEntry(entity.IdxSummaryKey, key4),
				index.Count(&i),
			))
			if i != 0 {
				t.Fatal()
			}
			ce(selIndex(
				index.MatchEntry(entity.IdxSummaryKey, key3),
				index.Count(&i),
			))
			if i != 1 {
				t.Fatal()
			}
		}))

		// ToContens with tx
		var key Key
		scope.Sub(
			func() filebase.ToContentsWithTx {
				return func(fn any) {
					ce(tx(fn))
				}
			},
		).Call(func(
			toContents filebase.ToContents,
		) {
			r := strings.NewReader("foo")
			keys, _, err := toContents(r, int64(r.Len()))
			ce(err)
			if len(keys) != 1 {
				t.Fatal()
			}
			if keys[0].String() != "c-entity:6f893e6ca357461a27335071b90f8781acf3c59dfe0a6667b6f6ca8ae389c7ee" {
				t.Fatal()
			}
			key = keys[0]

			ce(tx(func(
				fetch entity.Fetch,
			) {
				for _, key := range keys {
					var data []byte
					ce(fetch(key, &data))
				}
			}))

		})

		scope.Sub(
			func() Store {
				return kv
			},
		).Call(func(
			fetch entity.Fetch,
		) {
			var o any
			ce(fetch(key, &o))
			bs, ok := o.([]byte)
			if !ok {
				t.Fatal()
			}
			if string(bs) != "foo" {
				t.Fatal()
			}
		})

	})

}

func TestPebbleTxEntityDelete(
	t *testing.T,
	newPeb storepebble.New,
	newKV storekv.New,
	scope Scope,
) {
	defer he(nil, e4.TestingFatal(t))

	dir := t.TempDir()
	peb, err := newPeb(nil, dir)
	ce(err)
	defer peb.Close()

	scope.Sub(
		func() KVToStore {
			return func(kv storekv.KV) (store.Store, error) {
				return newKV(kv, "foo")
			}
		},
		UsePebbleTx,
		&peb,
	).Call(func(
		tx PebbleTx,
		scope Scope,
	) {

		ce(tx(func(
			save entity.SaveEntity,
			sel index.SelectIndex,
			del entity.Delete,
		) {

			s, err := save(42)
			ce(err)
			_ = s

			var c int
			ce(sel(
				index.MatchEntry(entity.IdxPairObjectSummary, s.Key),
				index.Count(&c),
			))
			if c != 1 {
				t.Fatal()
			}

			ce(del(s.Key))

			ce(sel(
				index.MatchEntry(entity.IdxPairObjectSummary, s.Key),
				index.Count(&c),
			))
			if c != 0 {
				t.Fatal()
			}

		}))

	})

}

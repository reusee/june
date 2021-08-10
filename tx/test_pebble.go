// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package tx

import (
	"errors"
	"fmt"
	"testing"

	"github.com/reusee/e4"
	"github.com/reusee/june/entity"
	"github.com/reusee/june/index"
	"github.com/reusee/june/store"
	"github.com/reusee/june/storekv"
	"github.com/reusee/june/storepebble"
	"github.com/reusee/pr"
)

func TestPebbleTx(
	t *testing.T,
	wt *pr.WaitTree,
	newPeb storepebble.New,
	newKV storekv.New,
	scope Scope,
) {
	defer he(nil, e4.TestingFatal(t))

	dir := t.TempDir()
	peb, err := newPeb(wt, nil, dir)
	ce(err)

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
		ce(tx(wt, func(
			save entity.Save,
		) {
			summary, err := save(entity.NSEntity, 42)
			ce(err)
			key1 = summary.Key
		}))

		ce(tx(wt, func(
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
		err = tx(wt, func(
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

		ce(tx(wt, func(
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
		err = tx(wt, func(
			save entity.Save,
			store store.Store,
		) {
			ce(tx(wt, func(
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

		ce(tx(wt, func(
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

	})

}

func TestPebbleTxEntityDelete(
	t *testing.T,
	wt *pr.WaitTree,
	newPeb storepebble.New,
	newKV storekv.New,
	scope Scope,
) {
	defer he(nil, e4.TestingFatal(t))

	dir := t.TempDir()
	peb, err := newPeb(wt, nil, dir)
	ce(err)

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

		ce(tx(wt, func(
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

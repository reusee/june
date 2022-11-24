// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package tx

import (
	"errors"
	"fmt"
	"testing"

	"github.com/reusee/e5"
	"github.com/reusee/june/entity"
	"github.com/reusee/june/index"
	"github.com/reusee/june/store"
	"github.com/reusee/june/storekv"
	"github.com/reusee/june/storepebble"
	"github.com/reusee/pr2"
)

func TestPebbleTx(
	t *testing.T,
	newPeb storepebble.New,
	newKV storekv.New,
	scope Scope,
	wg *pr2.WaitGroup,
) {
	defer he(nil, e5.TestingFatal(t))

	dir := t.TempDir()
	peb, err := newPeb(wg, nil, dir)
	ce(err)

	scope.Fork(
		func() KVToStore {
			return func(kv storekv.KV) (store.Store, error) {
				return newKV(wg, kv, "foo")
			}
		},
		UsePebbleTx,
		&peb,
	).Call(func(
		tx PebbleTx,
	) {

		// commit tx
		var key1 Key
		ce(tx(wg, func(
			save entity.Save,
		) {
			summary, err := save(wg, entity.NSEntity, 42)
			ce(err)
			key1 = summary.Key
		}))

		ce(tx(wg, func(
			fetch entity.Fetch,
			selIndex index.SelectIndex,
		) {
			var i int
			ce(fetch(key1, &i))
			if i != 42 {
				t.Fatal()
			}

			ce(selIndex(
				wg,
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
		err = tx(wg, func(
			save entity.Save,
		) {
			summary, err := save(wg, entity.NSEntity, 1)
			ce(err)
			key2 = summary.Key
			ce(errFoo)
		})
		if !errors.Is(err, errFoo) {
			t.Fatal()
		}

		ce(tx(wg, func(
			fetch entity.Fetch,
			selIndex index.SelectIndex,
		) {
			var i int
			err := fetch(key2, &i)
			if !errors.Is(err, store.ErrKeyNotFound) {
				t.Fatal()
			}

			ce(selIndex(
				wg,
				index.MatchEntry(entity.IdxSummaryKey, key2),
				index.Count(&i),
			))
			if i != 0 {
				t.Fatal()
			}
			ce(selIndex(
				wg,
				index.MatchEntry(entity.IdxSummaryKey, key1),
				index.Count(&i),
			))
			if i != 1 {
				t.Fatal()
			}
		}))

		// tx inside tx, partial commit
		var key3, key4 Key
		err = tx(wg, func(
			save entity.Save,
			store store.Store,
		) {
			ce(tx(wg, func(
				save entity.Save,
			) {
				summary, err := save(wg, entity.NSEntity, 99)
				ce(err)
				key3 = summary.Key
			}))

			// should see committed key
			ok, err := store.Exists(key3)
			ce(err)
			if !ok {
				t.Fatal()
			}

			summary, err := save(wg, entity.NSEntity, 1)
			ce(err)
			key4 = summary.Key
			ce(errFoo)
		})
		if !errors.Is(err, errFoo) {
			t.Fatal()
		}

		ce(tx(wg, func(
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
				wg,
				index.MatchEntry(entity.IdxSummaryKey, key4),
				index.Count(&i),
			))
			if i != 0 {
				t.Fatal()
			}
			ce(selIndex(
				wg,
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
	newPeb storepebble.New,
	newKV storekv.New,
	scope Scope,
	wg *pr2.WaitGroup,
) {
	defer he(nil, e5.TestingFatal(t))

	dir := t.TempDir()
	peb, err := newPeb(wg, nil, dir)
	ce(err)

	scope.Fork(
		func() KVToStore {
			return func(kv storekv.KV) (store.Store, error) {
				return newKV(wg, kv, "foo")
			}
		},
		UsePebbleTx,
		&peb,
	).Call(func(
		tx PebbleTx,
	) {

		ce(tx(wg, func(
			save entity.SaveEntity,
			sel index.SelectIndex,
			del entity.Delete,
		) {

			s, err := save(wg, 42)
			ce(err)
			_ = s

			var c int
			ce(sel(
				wg,
				index.MatchEntry(entity.IdxPairObjectSummary, s.Key),
				index.Count(&c),
			))
			if c != 1 {
				t.Fatal()
			}

			ce(del(wg, s.Key))

			ce(sel(
				wg,
				index.MatchEntry(entity.IdxPairObjectSummary, s.Key),
				index.Count(&c),
			))
			if c != 0 {
				t.Fatal()
			}

		}))

	})

}

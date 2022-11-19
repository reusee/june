// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package entity

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/reusee/e5"
	"github.com/reusee/june/index"
	"github.com/reusee/june/opts"
	"github.com/reusee/june/store"
	"github.com/reusee/sb"
)

func TestIndex(
	t *testing.T,
	store store.Store,
	save SaveEntity,
	fetch Fetch,
	checkRef CheckRef,
	cleanIndex CleanIndex,
	index Index,
	indexGC IndexGC,
) {
	defer he(nil, e5.TestingFatal(t))
	ctx := context.Background()

	// save
	var keys []Key
	for i := 0; i < 8; i++ {
		if summary, err := save(ctx, testIndex(i)); err != nil {
			t.Fatal(err)
		} else {
			keys = append(keys, summary.Key)
		}
	}

	// select
	var n int64
	nTokens := 0
	lenTuple := 0
	nKey := 0
	ce(Select(
		ctx,
		index,
		MatchEntry(idxTest{}),
		Tap(func(i int, key Key) {
			n++
			var value testIndex
			ce(fetch(ctx, key, &value))
			if int(value) != i {
				t.Fatal()
			}
		}),
		TapTokens(func(tokens sb.Tokens) {
			nTokens = len(tokens)
		}),
		IndexTapEntry(func(e IndexEntry) {
			lenTuple = len(e.Tuple)
		}),
		TapKey(func(_ Key) {
			nKey++
		}),
	))
	if n != 8 {
		t.Fatal()
	}
	if nTokens != 10 {
		t.Fatal()
	}
	if lenTuple != 1 {
		t.Fatalf("got %d\n", lenTuple)
	}
	if nKey != 8 {
		t.Fatal()
	}

	// check
	ce(checkRef(ctx))

	// delete
	ce(store.Delete(ctx, keys[:1]))

	// clean index
	n = 0
	m := 0
	nKeys := 0
	ce(cleanIndex(
		ctx,
		TapDeleteIndex(func(_ IndexEntry) {
			n++
		}),
		opts.TapInvalidKey(func(key Key) {
			m++
			if key != keys[0] {
				t.Fatal()
			}
		}),
		opts.TapKey(func(_ Key) {
			nKeys++
		}),
	))
	if m != 7 {
		t.Fatalf("got %d", m)
	}
	if n != 7 {
		t.Fatalf("got %d", n)
	}
	if nKeys != 88 {
		t.Fatalf("got %d", nKeys)
	}

	// save
	var summaryKeys []Key
	for i := 0; i < 8; i++ {
		if _, err := save(
			ctx,
			testIndex(i),
			SaveSummaryOptions([]SaveSummaryOption{
				TapKey(func(key Key) {
					summaryKeys = append(summaryKeys, key)
				}),
			}),
		); err != nil {
			t.Fatal(err)
		}
	}
	// delete summary
	ce(store.Delete(ctx, summaryKeys[:1]))
	// gc
	n = 0
	ce(indexGC(
		ctx,
		TapDeleteIndex(func(_ IndexEntry) {
			atomic.AddInt64(&n, 1)
		}),
	))
	if n != 7 {
		t.Fatalf("got %d", n)
	}

}

type idxTest struct {
	I int
}

var IdxTest = idxTest{}

func init() {
	index.Register(idxTest{})
}

type testIndex int

var _ HasIndex = testIndex(0)

func (t testIndex) EntityIndexes() (IndexSet, int64, error) {
	return IndexSet{
		NewEntry(idxTest{}, int(t)),
	}, 1, nil
}

func TestEmbeddingSameObject(
	t *testing.T,
	save SaveEntity,
	del Delete,
	sel index.SelectIndex,
) {
	defer he(nil, e5.TestingFatal(t))
	ctx := context.Background()

	type Foo struct {
		N int
		I testIndex
	}

	a := Foo{
		N: 1,
		I: testIndex(42),
	}
	s, err := save(ctx, a)
	ce(err)
	key1 := s.Key

	b := Foo{
		N: 2,
		I: testIndex(42),
	}
	s, err = save(ctx, b)
	ce(err)
	key2 := s.Key

	var n int
	ce(sel(
		ctx,
		MatchEntry(idxTest{}, 42),
		Count(&n),
	))
	if n != 1 {
		t.Fatal()
	}

	ce(del(ctx, key1))

	ce(sel(
		ctx,
		MatchEntry(idxTest{}, 42),
		Count(&n),
	))
	if n != 1 {
		t.Fatalf("got %d\n", n)
	}

	ce(del(ctx, key2))

	ce(sel(
		ctx,
		MatchEntry(idxTest{}, 42),
		Count(&n),
	))
	if n != 0 {
		t.Fatalf("got %d\n", n)
	}

}

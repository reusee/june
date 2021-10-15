// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package entity

import (
	"fmt"
	"hash/fnv"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/reusee/e4"
	"github.com/reusee/june/key"
	"github.com/reusee/june/store"
	"github.com/reusee/june/storemem"
	"github.com/reusee/june/storetap"
	"github.com/reusee/pp"
	"github.com/reusee/pr"
	"github.com/reusee/sb"
)

type testSaveFoo int

var _ HasIndex = testSaveFoo(0)

func (t testSaveFoo) EntityIndexes() (IndexSet, int64, error) {
	return IndexSet{
		IdxName(Name("foo")),
	}, 1, nil
}

func TestSave(
	t *testing.T,
	wt *pr.WaitTree,
	save SaveEntity,
	store store.Store,
	rebuildIndex RebuildIndex,
	updateIndex UpdateIndex,
	checkRef CheckRef,
	cleanIndex CleanIndex,
	newMemStore storemem.New,
	scope Scope,
	newHashState NewHashState,
	index Index,
	resave Resave,
	indexGC IndexGC,
) {
	defer he(nil, e4.TestingFatal(t))

	var err error
	var summary *Summary
	var bytesWritten, keysWritten int64
	scope.Fork(func(
		newTap storetap.New,
	) Store {
		return newTap(store, storetap.Funcs{
			Write: func(
				_ key.Namespace,
				_ sb.Stream,
				_ []WriteOption,
				res WriteResult,
				err error,
			) {
				keysWritten++
				bytesWritten += res.BytesWritten
			},
		})
	}).Call(func(
		save SaveEntity,
	) {
		summary, err = save(testSaveFoo(42))
		ce(err)
		if keysWritten != 2 {
			t.Fatalf("got %d", keysWritten)
		}
		if bytesWritten != 181 {
			t.Fatalf("got %d", bytesWritten)
		}
	})

	if !summary.Valid() {
		t.Fatal()
	}
	if summary.Key.Hash.String() != "151a3a0b4c88483512fc484d0badfedf80013ebb18df498bbee89ac5b69d7222" {
		t.Fatalf("got %x\n", summary.Key)
	}
	n := 0
	ce(summary.iterAll(
		func(path []Key, s Summary) error {
			n++
			if len(path) != 1 {
				t.Fatal()
			}
			if path[0].Hash.String() != "151a3a0b4c88483512fc484d0badfedf80013ebb18df498bbee89ac5b69d7222" {
				t.Fatalf("got %x\n", path[0])
			}
			return nil
		},
	))
	if n != 1 {
		t.Fatal()
	}

	var summaryHash []byte
	ce(sb.Copy(
		sb.Marshal(summary),
		sb.Hash(fnv.New128, &summaryHash, nil),
	))
	if fmt.Sprintf("%x", summaryHash) != "5e56b85fe057fb10ebb4d288e12144b3" {
		t.Fatalf("got %x", summaryHash)
	}

	numIdx := 0
	numType := 0
	numTypeFoo := 0
	numName := 0
	ce(Select(index, Unmarshal(func(tuple ...any) {
		numIdx++
		prefix, ok := tuple[0].(string)
		if !ok {
			// non Entry
			return
		}
		if strings.HasSuffix(prefix, "idxType") {
			numType++
			if tuple[1].(string) == "entity.testSaveFoo" {
				numTypeFoo++
				var key Key
				ce(sb.Copy(
					sb.Marshal(tuple[2]),
					sb.Unmarshal(&key),
				))
				if key != summary.Key {
					t.Fatal()
				}
			}
		}
		if strings.HasSuffix(prefix, "idxName") {
			numName++
			if tuple[1].(string) != "foo" {
				t.Fatal()
			}
		}
	})))

	if numIdx != 14 {
		t.Fatalf("got %d", numIdx)
	}
	if numType != 1 {
		t.Fatal()
	}
	if numTypeFoo != 1 {
		t.Fatal()
	}
	if numName != 1 {
		t.Fatal()
	}

	// pre index
	var savedVersion int64
	ce(Select(index,
		MatchPreEntry(summary.Key, IdxVersion),
		TapPre(func(v int64) {
			savedVersion = v
		}),
	))
	if savedVersion != 1 {
		t.Fatal()
	}

	t.Run("same data", func(t *testing.T) {
		defer he(nil, e4.TestingFatal(t))
		summary, err = save(testSaveFoo(42))
		ce(err)

		if !summary.Valid() {
			t.Fatal()
		}
		if summary.Key.Hash.String() != "151a3a0b4c88483512fc484d0badfedf80013ebb18df498bbee89ac5b69d7222" {
			t.Fatalf("got %x", summary.Key)
		}
		n := 0
		ce(summary.iterAll(
			func(path []Key, s Summary) error {
				n++
				if len(path) != 1 {
					t.Fatal()
				}
				if path[0].Hash.String() != "151a3a0b4c88483512fc484d0badfedf80013ebb18df498bbee89ac5b69d7222" {
					t.Fatal()
				}
				return nil
			},
		))
		if n != 1 {
			t.Fatal()
		}

		numIdx = 0
		numType = 0
		numTypeFoo = 0
		numName = 0
		ce(Select(index, Unmarshal(func(tuple ...any) {
			numIdx++
			prefix, ok := tuple[0].(string)
			if !ok {
				// non Entry
				return
			}
			if strings.HasSuffix(prefix, "idxType") {
				numType++
				if tuple[1].(string) == "entity.testSaveFoo" {
					numTypeFoo++
				}
			}
			if strings.HasSuffix(prefix, "idxName") {
				numName++
				if tuple[1].(string) != "foo" {
					t.Fatal()
				}
			}
		})))

		// no new
		if numIdx != 14 {
			t.Fatalf("got %d", numIdx)
		}
		if numType != 1 {
			t.Fatal()
		}
		if numTypeFoo != 1 {
			t.Fatal()
		}
		if numName != 1 {
			t.Fatal()
		}
	})

	t.Run("new data", func(t *testing.T) {
		defer he(nil, e4.TestingFatal(t))
		summary, err = save(testSaveFoo(43))
		ce(err)

		if !summary.Valid() {
			t.Fatal()
		}
		if summary.Key.Hash.String() != "37ec9616c907e0aa5e75ab7dfd7335c84ad77b2ea26fab8bd4b5da270b25ac9a" {
			t.Fatalf("got %x", summary.Key)
		}
		n := 0
		ce(summary.iterAll(
			func(path []Key, s Summary) error {
				n++
				if len(path) != 1 {
					t.Fatal()
				}
				if path[0].Hash.String() != "37ec9616c907e0aa5e75ab7dfd7335c84ad77b2ea26fab8bd4b5da270b25ac9a" {
					t.Fatal()
				}
				return nil
			},
		))
		if n != 1 {
			t.Fatal()
		}

		numIdx = 0
		numType = 0
		numTypeFoo = 0
		numName = 0
		ce(Select(
			index,
			Desc,
			Unmarshal(func(tuple ...any) {
				numIdx++
				prefix, ok := tuple[0].(string)
				if !ok {
					return
				}
				if strings.HasSuffix(prefix, "idxType") {
					numType++
					if tuple[1].(string) == "entity.testSaveFoo" {
						numTypeFoo++
					}
				}
				if strings.HasSuffix(prefix, "idxName") {
					numName++
					if tuple[1].(string) != "foo" {
						t.Fatal()
					}
				}
			}),
		))

		if numIdx != 28 {
			t.Fatalf("got %d", numIdx)
		}
		if numType != 2 {
			t.Fatalf("got %d", numType)
		}
		if numTypeFoo != 2 {
			t.Fatal()
		}
		if numName != 2 {
			t.Fatal()
		}
	})

	t.Run("reindex", func(t *testing.T) {
		defer he(nil, e4.TestingFatal(t))
		var before int
		ce(Select(
			index,
			Count(&before),
		))

		var num int64
		if n, err := rebuildIndex(
			WithIndexSaveOptions([]IndexSaveOption{
				IndexTapEntry(func(entry IndexEntry) {
					atomic.AddInt64(&num, 1)
				}),
			}),
		); err != nil {
			t.Fatal(err)
		} else if n != 2 {
			t.Fatalf("got %d\n", n)
		}

		if num != 14 {
			t.Fatalf("got %d", num)
		}

		var after int
		ce(Select(
			index,
			Count(&after),
		))
		if after != before {
			t.Fatal()
		}

		if n, err := updateIndex(); err != nil {
			t.Fatal(err)
		} else if n != 0 {
			t.Fatalf("got %d\n", n)
		}

	})

	t.Run("index from summary", func(t *testing.T) {
		defer he(nil, e4.TestingFatal(t))

		iter, closer, err := index.Iter(
			nil,
			nil,
			Asc,
		)
		ce(err)
		defer closer.Close()
		indexes := make(map[Hash]IndexEntry)
		n := 0
		ce(pp.Copy(iter, pp.Tap(func(v any) (err error) {
			s := v.(sb.Stream)
			defer he(&err)
			var entry *IndexEntry
			var preEntry *IndexPreEntry
			var h []byte
			ce(sb.Copy(
				s,
				sb.AltSink(
					sb.Unmarshal(&entry),
					sb.Unmarshal(&preEntry),
				),
				sb.Hash(newHashState, &h, nil),
			))
			var hash Hash
			copy(hash[:], h)
			if entry != nil {
				indexes[hash] = *entry
			}
			n++
			return nil
		})))

		newIndexes := make(map[Hash]IndexEntry)
		var n2 int
		manager := IndexManager(newMemStore(wt))
		storeID, err := store.ID()
		ce(err)
		scope.Fork(
			&storeID,
			&manager,
		).Call(func(
			store Store,
			index Index,
			fetch Fetch,
			saveSummary SaveSummary,
		) {

			ce(store.IterKeys(NSSummary, func(key Key) (err error) {
				defer he(&err)
				var summary Summary
				ce(fetch(key, &summary))
				ce(saveSummary(&summary, false))
				return nil
			}))

			iter, closer, err := index.Iter(
				nil,
				nil,
				Asc,
			)
			ce(err)
			defer closer.Close()
			ce(pp.Copy(iter, pp.Tap(func(v any) error {
				s := v.(sb.Stream)
				var entry *IndexEntry
				var preEntry *IndexPreEntry
				var h []byte
				if err := sb.Copy(
					s,
					sb.AltSink(
						sb.Unmarshal(&entry),
						sb.Unmarshal(&preEntry),
					),
					sb.Hash(newHashState, &h, nil),
				); err != nil {
					t.Fatal(err)
				}
				var hash Hash
				copy(hash[:], h)
				if entry != nil {
					newIndexes[hash] = *entry
				}
				n2++
				return nil
			})))

		})

		if len(indexes) != len(newIndexes) {
			t.Fatal()
		}
		if n != n2 {
			t.Fatal()
		}

	})

	t.Run("index gc", func(t *testing.T) {
		defer he(nil, e4.TestingFatal(t))
		ce(indexGC())
	})

	t.Run("index clean", func(t *testing.T) {
		defer he(nil, e4.TestingFatal(t))
		ce(cleanIndex())
	})

	t.Run("resave", func(t *testing.T) {
		var n int64
		var m int64
		ce(resave(
			[]any{
				testSaveFoo(0),
			},
			TapKey(func(_ Key) {
				atomic.AddInt64(&n, 1)
			}),
			SaveOptions{
				TapSummary(func(_ *Summary) {
					atomic.AddInt64(&m, 1)
				}),
			},
		))
		if n != 2 {
			t.Fatalf("got %d\n", n)
		}
		if m != 2 {
			t.Fatal()
		}
	})

	ce(checkRef())

}

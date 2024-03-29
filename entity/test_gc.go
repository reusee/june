// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package entity

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/reusee/e5"
	"github.com/reusee/june/index"
	"github.com/reusee/june/store"
	"github.com/reusee/june/storemem"
	"github.com/reusee/pr2"
)

func TestGC(
	t *testing.T,
	store store.Store,
	saveEntity SaveEntity,
	fetch Fetch,
	gc GC,
	checkRef CheckRef,
	index Index,
	deleteSummary DeleteSummary,
	wg *pr2.WaitGroup,
) {
	defer he(nil, e5.TestingFatal(t))

	type Root struct {
		Key Key
	}

	type Foo struct {
		Content any
	}

	var fns []func() Key
	var keys []Key
	fns = []func() Key{

		// plain object
		func() Key {
			s, err := saveEntity(wg, Foo{
				Content: fmt.Sprintf("%d", rand.Int63()),
			})
			ce(err)
			keys = append(keys, s.Key)
			return s.Key
		},

		// nested
		func() Key {
			key := fns[rand.Intn(len(fns))]()
			s, err := saveEntity(wg, Foo{
				Content: key,
			})
			ce(err)
			keys = append(keys, s.Key)
			return s.Key
		},

		// existed
		func() Key {
			var key Key
			if len(keys) > 0 {
				key = keys[rand.Intn(len(keys))]
			} else {
				key = fns[rand.Intn(len(fns))]()
			}
			s, err := saveEntity(wg, Foo{
				Content: key,
			})
			ce(err)
			keys = append(keys, s.Key)
			return s.Key
		},
	}

	rootKeys := make(map[Key]bool)
	for i := 0; i < 50; i++ {
		s, err := saveEntity(wg, Root{
			Key: fns[rand.Intn(len(fns))](),
		})
		ce(err)
		key := s.Key
		rootKeys[key] = true
	}

	numDead := 0

	for len(rootKeys) > 1 {
		var key Key
		for k := range rootKeys {
			key = k
			delete(rootKeys, k)
			break
		}

		// delete
		var toDelete []Key
		ce(Select(
			index,
			MatchEntry(IdxSummaryKey, key),
			Tap(func(_ Key, summaryKey Key) {
				toDelete = append(toDelete, summaryKey)
			}),
		))
		deleted := false
		for _, summaryKey := range toDelete {
			deleted = true
			var summary Summary
			ce((fetch(summaryKey, &summary)))
			ce(deleteSummary(wg, &summary, summaryKey))
		}
		if deleted {
			ce(store.Delete([]Key{key}))
		}

		// gc
		var keys []Key
		for key := range rootKeys {
			keys = append(keys, key)
		}
		var numMarked, numReachable, numItered, numDeadObjs, numSweeped int64
		ce(gc(
			wg,
			keys,
			TapMarkKey(func(_ Key) {
				atomic.AddInt64(&numMarked, 1)
			}),
			TapReachableObjects(func(reachable *sync.Map) {
				reachable.Range(func(_, _ any) bool {
					numReachable++
					return true
				})
			}),
			TapIterKey(func(_ Key) {
				atomic.AddInt64(&numItered, 1)
			}),
			TapDeadObjects(func(deadObjs []DeadObject) {
				numDeadObjs = int64(len(deadObjs))
				numDead += len(deadObjs)
			}),
			TapSweepDeadObject(func(_ DeadObject) {
				atomic.AddInt64(&numSweeped, 1)
			}),
		))

		if numMarked == 0 {
			t.Fatal()
		}
		if numReachable <= 0 || numReachable > numMarked {
			t.Fatal()
		}
		if numItered < numMarked {
			t.Fatal()
		}
		if numDeadObjs != numSweeped {
			t.Fatal()
		}

		// check
		ce(checkRef(wg))

	}

	if numDead == 0 {
		t.Fatal()
	}

}

func TestGCWithEmptyIndex(
	t *testing.T,
	save Save,
	newMemStore storemem.New,
	scope Scope,
	gc GC,
	wg *pr2.WaitGroup,
) {
	defer he(nil, e5.TestingFatal(t))

	res, err := save(wg, NSEntity, 42)
	ce(err)
	type Foo struct {
		Key Key
	}
	res, err = save(wg, NSEntity, Foo{
		Key: res.Key,
	})
	ce(err)

	n := 0
	ce(gc(wg, []Key{
		res.Key,
	}, TapSweepDeadObject(func(_ DeadObject) {
		n++
	})))
	if n > 0 {
		t.Fatal()
	}

	indexManager := newMemStore(wg)
	scope.Fork(func() index.IndexManager {
		return indexManager
	}).Call(func(
		gc GC,
	) {
		n := 0
		ce(gc(wg, []Key{
			res.Key,
		}, TapSweepDeadObject(func(_ DeadObject) {
			n++
		})))
		if n > 0 {
			t.Fatal()
		}
	})

}

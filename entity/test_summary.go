// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package entity

import (
	"fmt"
	"hash/fnv"
	"testing"

	"github.com/reusee/e4"
	"github.com/reusee/june/opts"
	"github.com/reusee/sb"
)

func TestSummary(
	t *testing.T,
	save SaveEntity,
) {
	defer he(nil, e4.TestingFatal(t))

	summary, err := save(testSummary{
		I:  42,
		S:  "foo",
		Is: []int{1, 2, 3},
		M: map[int]int{
			1: 2,
			3: 4,
			5: 6,
		},
	})
	ce(err)

	var hash []byte
	ce(sb.Copy(
		sb.Marshal(summary),
		sb.Hash(fnv.New128, &hash, nil),
	))
	if fmt.Sprintf("%x", hash) != "a5e73de8aed7a3cbd99f4a12afc7ec53" {
		t.Fatalf("got %x", hash)
	}

	var summary2 Summary
	ce(sb.Copy(
		sb.Marshal(summary),
		sb.Unmarshal(&summary2),
	))

	var hash2 []byte
	ce(sb.Copy(
		sb.Marshal(summary2),
		sb.Hash(fnv.New128, &hash2, nil),
	))
	if fmt.Sprintf("%x", hash2) != "a5e73de8aed7a3cbd99f4a12afc7ec53" {
		t.Fatalf("got %x", hash2)
	}

}

type testSummary struct {
	I  int
	S  string
	Is []int
	M  map[int]int
}

type testSummaryUpdateFoo struct {
	indexes func() IndexSet
	I       int
}

var _ HasIndex = testSummaryUpdateFoo{}

func (t testSummaryUpdateFoo) EntityIndexes() (IndexSet, int64, error) {
	return t.indexes(), 1, nil
}

func TestSummaryUpdate(
	t *testing.T,
	save Save,
	fetch Fetch,
	checkRef CheckRef,
	cleanIndex CleanIndex,
	indexGC IndexGC,
) {
	defer he(nil, e4.TestingFatal(t))

	var summaryKey1 Key
	summary1, err := save(
		NSEntity, testSummaryUpdateFoo{
			I: 42,
			indexes: func() IndexSet {
				return IndexSet{
					NewEntry(IdxTest, 1),
				}
			},
		},
		SaveSummaryOptions{
			TapKey(func(key Key) {
				summaryKey1 = key
			}),
		},
	)
	ce(err)

	summary2, err := save(NSEntity, testSummaryUpdateFoo{
		I: 42,
		indexes: func() IndexSet {
			return IndexSet{
				NewEntry(IdxTest, 1),
				NewEntry(IdxTest, 2),
			}
		},
	})
	ce(err)

	// same entity
	if summary1.Key != summary2.Key {
		t.Fatal()
	}

	// deleted
	err = fetch(summaryKey1, &summary1)
	if !is(err, ErrKeyNotFound) {
		t.Fatal()
	}

	// check ref
	ce(checkRef())

	// no invalid index
	n := 0
	ce(cleanIndex(opts.TapInvalidKey(func(_ Key) {
		n++
	})))
	if n > 0 {
		t.Fatal()
	}

	// no garbage index
	n = 0
	ce(indexGC(TapDeleteIndex(func(_ IndexEntry) {
		n++
	})))
	if n > 0 {
		t.Fatal()
	}

}

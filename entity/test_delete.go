// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package entity

import (
	"testing"

	"github.com/reusee/e5"
	"github.com/reusee/june/index"
)

func TestDelete(
	t *testing.T,
	store Store,
	save Save,
	del Delete,
	checkRef CheckRef,
	sel index.SelectIndex,
) {
	defer he(nil, e5.TestingFatal(t))

	var summary1 *Summary
	var summaryKey1 Key
	res, err := save(NSEntity, 42,
		TapSummary(func(s *Summary) {
			summary1 = s
		}),
		SaveSummaryOptions{
			TapKey(func(key Key) {
				summaryKey1 = key
			}),
		},
	)
	ce(err)
	key1 := res.Key
	_ = summary1
	_ = summaryKey1

	type Foo struct {
		Key Key
	}
	var summary2 *Summary
	var summaryKey2 Key
	res, err = save(NSEntity, Foo{Key: key1},
		TapSummary(func(s *Summary) {
			summary2 = s
		}),
		SaveSummaryOptions{
			TapKey(func(key Key) {
				summaryKey2 = key
			}),
		},
	)
	ce(err)
	key2 := res.Key
	_ = summary2

	err = del(key1)
	var beingRefered *BeingRefered
	if !as(err, &beingRefered) {
		t.Fatal()
	}
	if beingRefered.Key != key1 {
		t.Fatal()
	}
	if beingRefered.Key.Namespace != NSEntity {
		t.Fatal()
	}
	if len(beingRefered.By) != 1 {
		t.Fatal()
	}
	if _, ok := beingRefered.By[key2]; !ok {
		t.Fatal()
	}

	ce(del(summaryKey2))
	ce(checkRef())

	ce(del(key1))
	ce(checkRef())

	// check indexes
	ce(sel(
		MatchType(int(0)),
		Tap(func(_ string, key Key) {
			if key == key1 {
				t.Fatalf("IdxType not deleted")
			}
		}),
	))

}

// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package entity

import (
	"testing"

	"github.com/reusee/e5"
	"github.com/reusee/june/index"
	"github.com/reusee/pr2"
)

func TestDelete(
	t *testing.T,
	save Save,
	del Delete,
	checkRef CheckRef,
	sel index.SelectIndex,
	wg *pr2.WaitGroup,
) {
	defer he(nil, e5.TestingFatal(t))

	var summary1 *Summary
	var summaryKey1 Key
	res, err := save(wg, NSEntity, 42,
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
	res, err = save(wg, NSEntity, Foo{Key: key1},
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

	err = del(wg, key1)
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

	ce(del(wg, summaryKey2))
	ce(checkRef(wg))

	ce(del(wg, key1))
	ce(checkRef(wg))

	// check indexes
	ce(sel(
		wg,
		MatchType(int(0)),
		Tap(func(_ string, key Key) {
			if key == key1 {
				t.Fatalf("IdxType not deleted")
			}
		}),
	))

}

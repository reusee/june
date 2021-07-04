// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package entity

import (
	"github.com/reusee/ling/v2/index"
)

type idxReferedBy struct {
	Key1 Key
}

var IdxReferedBy = idxReferedBy{}

var IdxPairReferredObject = idxReferedBy{}

func init() {
	index.Register(IdxReferedBy)
}

func (_ Def) IdxReferedByFuncs() (
	add OnSummaryIndexAdd,
	del OnSummaryIndexDelete,
) {

	add = func(
		summary *Summary,
		summaryKey Key,
	) (
		entries []IndexEntry,
		err error,
	) {
		defer he(&err)
		for _, referedKey := range summary.ReferedKeys {
			entries = append(entries,
				NewEntry(IdxReferedBy, referedKey, summary.Key),
			)
		}
		entries = append(entries,
			NewEntry(IdxReferedBy, summary.Key, summaryKey),
		)
		return
	}

	del = func(
		summary *Summary,
		summaryKey Key,
	) (
		entries []IndexEntry,
		err error,
	) {
		defer he(&err)
		for _, referedKey := range summary.ReferedKeys {
			entries = append(entries,
				NewEntry(IdxReferedBy, referedKey, summary.Key),
			)
		}
		entries = append(entries,
			NewEntry(IdxReferedBy, summary.Key, summaryKey),
		)
		return
	}

	return
}

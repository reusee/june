// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package entity

import (
	"github.com/reusee/ling/v2/index"
)

type idxSummaryOf struct {
	SummaryKey Key // key of summary
}

var IdxSummaryOf = idxSummaryOf{}

var IdxPairSummaryObject = idxSummaryOf{}

func init() {
	index.Register(IdxSummaryOf)
}

func (_ Def) IdxSummaryOfFuncs() (
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
		entries = append(entries,
			NewEntry(IdxSummaryOf, summaryKey, summary.Key),
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
		entries = append(entries,
			NewEntry(IdxSummaryOf, summaryKey, summary.Key),
		)
		return
	}

	return
}

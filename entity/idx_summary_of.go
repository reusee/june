// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package entity

import (
	"context"

	"github.com/reusee/june/index"
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
		ctx context.Context,
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
		ctx context.Context,
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

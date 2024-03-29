// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package entity

import (
	"context"

	"github.com/reusee/june/index"
)

type idxSummaryKey struct {
	Key Key // key of entity
}

var IdxSummaryKey = idxSummaryKey{}

var IdxPairObjectSummary = idxSummaryKey{}

func init() {
	index.Register(IdxSummaryKey)
}

func (Def) IdxSummaryKeyFuncs() (
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
			NewEntry(IdxSummaryKey, summary.Key, summaryKey),
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
			NewEntry(IdxSummaryKey, summary.Key, summaryKey),
		)
		return
	}

	return
}

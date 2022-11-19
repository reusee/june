// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package entity

import (
	"context"

	"github.com/reusee/june/index"
)

type idxReferTo struct {
	Key1 Key
}

var IdxReferTo = idxReferTo{}

var IdxPairObjectReferred = idxReferTo{}

func init() {
	index.Register(IdxReferTo)
}

func (_ Def) IdxReferToFuncs() (
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
		for _, referedKey := range summary.ReferedKeys {
			entries = append(entries,
				NewEntry(IdxReferTo, summary.Key, referedKey),
			)
		}
		entries = append(entries,
			NewEntry(IdxReferTo, summaryKey, summary.Key),
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
		for _, referedKey := range summary.ReferedKeys {
			entries = append(entries,
				NewEntry(IdxReferTo, summary.Key, referedKey),
			)
		}
		entries = append(entries,
			NewEntry(IdxReferTo, summaryKey, summary.Key),
		)
		return
	}

	return
}

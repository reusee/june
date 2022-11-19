// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package entity

import (
	"github.com/reusee/june/key"
	"github.com/reusee/sb"
)

type IndexSet []IndexEntry

type HashIndexSet func(
	set IndexSet,
) (
	Hash,
	error,
)

func (Def) HashIndexSet(
	newHashState key.NewHashState,
) HashIndexSet {
	return func(
		set IndexSet,
	) (
		hash Hash,
		err error,
	) {
		defer he(&err)
		var bs []byte
		ce(sb.Copy(
			sb.Marshal(set),
			sb.Hash(newHashState, &bs, nil),
		))
		copy(hash[:], bs)
		return
	}
}

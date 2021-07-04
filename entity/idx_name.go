// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package entity

import (
	"github.com/reusee/ling/v2/index"
)

// IdxName

type idxName func(
	Name Name,
) IndexEntry

var IdxName = idxName(func(name Name) IndexEntry {
	return NewEntry(idxName(nil), name)
})

func init() {
	index.Register(IdxName)
}

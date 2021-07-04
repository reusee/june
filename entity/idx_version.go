// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package entity

import (
	"github.com/reusee/june/index"
)

type idxVersion func(
	Version int64,
) IndexEntry

var IdxVersion = idxVersion(func(version int64) IndexEntry {
	return NewEntry(idxVersion(nil), version)
})

func init() {
	index.Register(IdxVersion)
}

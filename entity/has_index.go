// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package entity

import (
	"reflect"
)

//TODO UpgradeIndex

type HasIndex interface {
	EntityIndexes() (
		set IndexSet,
		version int64,
		err error,
	)
}

var hasIndexType = reflect.TypeOf((*HasIndex)(nil)).Elem()

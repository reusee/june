// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package entity

import (
	"reflect"

	"github.com/reusee/june/index"
	"github.com/reusee/june/naming"
)

type idxType func(
	Type string,
) IndexEntry

var IdxType = idxType(func(_type string) IndexEntry {
	return NewEntry(idxType(nil), _type)
})

func init() {
	index.Register(IdxType)
}

func MatchType(o any) index.Entry {
	typeName := naming.Type(reflect.TypeOf(o))
	return index.MatchEntry(IdxType, typeName)
}

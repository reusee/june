// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package index

import (
	"reflect"

	"github.com/reusee/june/naming"
)

type idxUnknown string

var nameOfIdxUnknown = naming.Type(
	reflect.TypeOf(idxUnknown("")),
)

var idxUnknownType = reflect.TypeOf((*idxUnknown)(nil)).Elem()

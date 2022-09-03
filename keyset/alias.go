// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package keyset

import (
	"fmt"

	"github.com/reusee/dscope"
	"github.com/reusee/e5"
	"github.com/reusee/june/key"
)

type (
	Scope = dscope.Scope
	Key   = key.Key
)

var (
	ce = e5.CheckWithStacktrace
	he = e5.Handle
	pt = fmt.Printf
)

// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package keyset

import (
	"fmt"

	"github.com/reusee/dscope"
	"github.com/reusee/june/juneerr"
	"github.com/reusee/june/key"
)

type (
	Scope = dscope.Scope
	Key   = key.Key
)

var (
	ce = juneerr.Check
	he = juneerr.Handle
	pt = fmt.Printf
)

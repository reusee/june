// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storesqlite

import (
	"fmt"

	"github.com/reusee/june/juneerr"
)

var (
	ce = juneerr.Check
	he = juneerr.Handle
	we = juneerr.Wrap
	pt = fmt.Printf
)

type (
	any = interface{}
)

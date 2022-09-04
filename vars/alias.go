// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package vars

import (
	"errors"
	"fmt"

	"github.com/reusee/dscope"
	"github.com/reusee/june/juneerr"
)

type (
	any   = interface{}
	Scope = dscope.Scope
)

var (
	is = errors.Is
	as = errors.As
	pt = fmt.Printf

	ce = juneerr.Check
	he = juneerr.Handle
	we = juneerr.Wrap
)

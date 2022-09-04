// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package filebase

import (
	"errors"
	"fmt"

	"github.com/reusee/dscope"
	"github.com/reusee/june/juneerr"
	"github.com/reusee/june/key"
)

type (
	any = interface{}

	Key = key.Key

	Scope = dscope.Scope
)

var (
	ce = juneerr.Check
	he = juneerr.Handle
	we = juneerr.Wrap
	is = errors.Is
	pt = fmt.Printf
)

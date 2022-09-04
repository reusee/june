// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package index

import (
	"errors"
	"fmt"

	"github.com/reusee/dscope"
	"github.com/reusee/june/juneerr"
	"github.com/reusee/june/key"
	"github.com/reusee/june/opts"
	"github.com/reusee/june/store"
	"github.com/reusee/pp"
)

var (
	pt = fmt.Printf
	ce = juneerr.Check
	he = juneerr.Handle
	we = juneerr.Wrap
	is = errors.Is
)

type (
	any = interface{}

	Scope = dscope.Scope

	StoreID = store.ID
	Store   = store.Store

	Key = key.Key

	TapKey = opts.TapKey

	Src = pp.Src
)

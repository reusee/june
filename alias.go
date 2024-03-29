// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package june

import (
	"fmt"

	"github.com/reusee/june/index"
	"github.com/reusee/june/juneerr"
	"github.com/reusee/june/store"
)

type (
	any = interface{}

	Store = store.Store

	IndexManager = index.IndexManager
)

var (
	pt = fmt.Printf
	sp = fmt.Sprintf
	ce = juneerr.Check
	he = juneerr.Handle
)

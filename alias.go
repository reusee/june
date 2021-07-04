// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package ling

import (
	"fmt"

	"github.com/reusee/e4"
	"github.com/reusee/ling/v2/index"
	"github.com/reusee/ling/v2/store"
)

type (
	any = interface{}

	Store = store.Store

	IndexManager = index.IndexManager
)

var (
	pt = fmt.Printf
	sp = fmt.Sprintf
	ce = e4.Check
	he = e4.Handle
)

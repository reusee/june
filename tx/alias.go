// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package tx

import (
	"fmt"

	"github.com/reusee/dscope"
	"github.com/reusee/e4"
	"github.com/reusee/ling/v2/index"
	"github.com/reusee/ling/v2/key"
	"github.com/reusee/ling/v2/store"
)

type (
	any = interface{}

	Scope        = dscope.Scope
	Store        = store.Store
	IndexManager = index.IndexManager
	Key          = key.Key
)

var (
	ce, he = e4.Check, e4.Handle
	pt     = fmt.Printf
)

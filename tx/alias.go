// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package tx

import (
	"fmt"

	"github.com/reusee/dscope"
	"github.com/reusee/e5"
	"github.com/reusee/june/index"
	"github.com/reusee/june/key"
	"github.com/reusee/june/store"
)

type (
	any = interface{}

	Scope        = dscope.Scope
	Store        = store.Store
	IndexManager = index.IndexManager
	Key          = key.Key
)

var (
	ce, he = e5.CheckWithStacktrace, e5.Handle
	pt     = fmt.Printf
)

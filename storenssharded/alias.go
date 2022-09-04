// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storenssharded

import (
	"errors"
	"fmt"

	"github.com/reusee/june/juneerr"
	"github.com/reusee/june/store"
)

type (
	any = interface{}

	Key         = store.Key
	StoreID     = store.ID
	WriteResult = store.WriteResult
	WriteOption = store.WriteOption
)

var (
	is = errors.Is
	pt = fmt.Printf
	he = juneerr.Handle
	ce = juneerr.Check
)

// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package stores3

import (
	"errors"
	"fmt"

	"github.com/reusee/dscope"
	"github.com/reusee/june/juneerr"
	"github.com/reusee/june/store"
)

var (
	is = errors.Is
	as = errors.As
	pt = fmt.Printf
	we = juneerr.Wrap
	ce = juneerr.Check
	he = juneerr.Handle

	Break = store.Break

	ErrKeyNotFound = store.ErrKeyNotFound
	ErrClosed      = store.ErrClosed
)

type (
	Scope = dscope.Scope
)

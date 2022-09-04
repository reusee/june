// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storeonedrive

import (
	"errors"
	"fmt"

	"github.com/reusee/june/juneerr"
	"github.com/reusee/june/store"
)

var (
	we = juneerr.Wrap
	ce = juneerr.Check
	he = juneerr.Handle
	pt = fmt.Printf
	as = errors.As
	is = errors.Is

	ErrClosed = store.ErrClosed
)

type (
	any = interface{}
)

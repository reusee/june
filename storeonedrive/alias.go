// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storeonedrive

import (
	"errors"
	"fmt"

	"github.com/reusee/e4"
	"github.com/reusee/ling/v2/store"
)

var (
	we = e4.Wrap
	ce = e4.Check
	he = e4.Handle
	pt = fmt.Printf
	as = errors.As
	is = errors.Is

	ErrClosed = store.ErrClosed
)

type (
	any = interface{}
)

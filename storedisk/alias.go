// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storedisk

import (
	"errors"
	"fmt"

	"github.com/reusee/june/codec"
	"github.com/reusee/june/fsys"
	"github.com/reusee/june/juneerr"
	"github.com/reusee/june/store"
)

type (
	any   = interface{}
	Codec = codec.Codec
)

var (
	is = errors.Is
	pt = fmt.Printf
	we = juneerr.Wrap
	he = juneerr.Handle
	ce = juneerr.Check

	Break = store.Break

	PathSeparator = fsys.PathSeparator

	ErrKeyNotFound = store.ErrKeyNotFound
	ErrClosed      = store.ErrClosed
)

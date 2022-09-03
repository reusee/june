// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storedisk

import (
	"errors"
	"fmt"

	"github.com/reusee/e5"
	"github.com/reusee/june/codec"
	"github.com/reusee/june/fsys"
	"github.com/reusee/june/store"
)

type (
	any   = interface{}
	Codec = codec.Codec
)

var (
	is = errors.Is
	pt = fmt.Printf
	we = e5.WrapWithStacktrace
	he = e5.Handle
	ce = e5.CheckWithStacktrace

	Break = store.Break

	PathSeparator = fsys.PathSeparator

	ErrKeyNotFound = store.ErrKeyNotFound
	ErrClosed      = store.ErrClosed
)

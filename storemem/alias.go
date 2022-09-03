// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storemem

import (
	"errors"
	"fmt"

	"github.com/reusee/e5"
	"github.com/reusee/june/codec"
	"github.com/reusee/june/index"
	"github.com/reusee/june/key"
	"github.com/reusee/june/store"
	"github.com/reusee/june/storekv"
)

type (
	any = interface{}

	Key           = key.Key
	Order         = index.Order
	Codec         = codec.Codec
	IndexEntry    = index.Entry
	StoreIndex    = index.StoreIndex
	IndexTapEntry = index.TapEntry

	StoreID = store.ID
)

var (
	is = errors.Is
	pt = fmt.Printf
	we = e5.WrapWithStacktrace
	ce = e5.CheckWithStacktrace
	he = e5.Handle

	Asc   = index.Asc
	Break = store.Break

	ErrKeyNotFound = storekv.ErrKeyNotFound
	ErrClosed      = store.ErrClosed
)

// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storemem

import (
	"errors"
	"fmt"

	"github.com/reusee/e4"
	"github.com/reusee/ling/v2/codec"
	"github.com/reusee/ling/v2/index"
	"github.com/reusee/ling/v2/key"
	"github.com/reusee/ling/v2/store"
	"github.com/reusee/ling/v2/storekv"
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
	we = e4.Wrap
	ce = e4.Check
	he = e4.Handle

	Asc   = index.Asc
	Break = store.Break

	ErrKeyNotFound = storekv.ErrKeyNotFound
	ErrClosed      = store.ErrClosed
)

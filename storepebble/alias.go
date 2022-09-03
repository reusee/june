// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storepebble

import (
	"errors"
	"fmt"

	"github.com/cockroachdb/pebble/vfs"
	"github.com/reusee/e5"
	"github.com/reusee/june/codec"
	"github.com/reusee/june/index"
	"github.com/reusee/june/key"
	"github.com/reusee/june/store"
)

type (
	any = interface{}

	Key             = key.Key
	Codec           = codec.Codec
	IndexEntry      = index.Entry
	StoreID         = store.ID
	StoreIndex      = index.StoreIndex
	IndexTapEntry   = index.TapEntry
	IndexSaveOption = index.SaveOption
)

var (
	is = errors.Is
	pt = fmt.Printf
	we = e5.WrapWithStacktrace
	ce = e5.CheckWithStacktrace
	he = e5.Handle

	Break = store.Break

	NewMemFS = vfs.NewMem

	ErrKeyNotFound = store.ErrKeyNotFound
	ErrClosed      = store.ErrClosed
)

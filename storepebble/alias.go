// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storepebble

import (
	"errors"
	"fmt"

	"github.com/cockroachdb/pebble/vfs"
	"github.com/reusee/e4"
	"github.com/reusee/ling/v2/codec"
	"github.com/reusee/ling/v2/index"
	"github.com/reusee/ling/v2/key"
	"github.com/reusee/ling/v2/store"
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
	we = e4.Wrap
	ce = e4.Check
	he = e4.Handle

	Break = store.Break

	NewMemFS = vfs.NewMem

	ErrKeyNotFound = store.ErrKeyNotFound
	ErrClosed      = store.ErrClosed
)

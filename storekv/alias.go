// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storekv

import (
	"errors"
	"fmt"

	"github.com/reusee/e4"
	"github.com/reusee/ling/v2/codec"
	"github.com/reusee/ling/v2/key"
	"github.com/reusee/ling/v2/opts"
	"github.com/reusee/ling/v2/store"
)

type (
	any = interface{}

	Key = key.Key

	Codec          = codec.Codec
	Cache          = store.Cache
	WriteOption    = store.WriteOption
	WriteResult    = store.WriteResult
	TapKey         = opts.TapKey
	TapWriteResult = store.TapWriteResult
	StoreID        = store.ID
)

var (
	is = errors.Is
	as = errors.As
	pt = fmt.Printf
	ce = e4.Check
	he = e4.Handle
	we = e4.Wrap

	Break = store.Break

	DefaultCodec = codec.DefaultCodec

	ErrKeyNotFound = store.ErrKeyNotFound
	ErrKeyNotMatch = store.ErrKeyNotMatch
	ErrRead        = store.ErrRead
)

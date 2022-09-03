// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storestacked

import (
	"errors"

	"github.com/reusee/e5"
	"github.com/reusee/june/key"
	"github.com/reusee/june/store"
)

type (
	any = interface{}

	Key         = key.Key
	WriteOption = store.WriteOption
	StoreID     = store.ID
)

var (
	is = errors.Is
	as = errors.As

	ErrIgnore = store.ErrIgnore
	Break     = store.Break

	ce = e5.CheckWithStacktrace
	he = e5.Handle
	we = e5.WrapWithStacktrace

	ErrKeyNotFound = store.ErrKeyNotFound
	ErrKeyNotMatch = store.ErrKeyNotMatch
	ErrClosed      = store.ErrClosed
)

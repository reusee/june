// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storestacked

import (
	"errors"

	"github.com/reusee/e4"
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

	ce = e4.Check
	he = e4.Handle
	we = e4.Wrap

	ErrKeyNotFound = store.ErrKeyNotFound
	ErrKeyNotMatch = store.ErrKeyNotMatch
	ErrClosed      = store.ErrClosed
)

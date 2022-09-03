// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package entity

import (
	"errors"
	"fmt"

	"github.com/reusee/dscope"
	"github.com/reusee/e5"
	"github.com/reusee/june/index"
	"github.com/reusee/june/key"
	"github.com/reusee/june/opts"
	"github.com/reusee/june/store"
)

type (
	any = interface{}

	Key           = key.Key
	Hash          = key.Hash
	KeyPath       = key.KeyPath
	WriteResult   = store.WriteResult
	WriteOption   = store.WriteOption
	Order         = index.Order
	Store         = store.Store
	Scope         = dscope.Scope
	IndexEntry    = index.Entry
	IndexPreEntry = index.PreEntry
	NewHashState  = key.NewHashState
	Limit         = index.Limit
	Index         = index.Index
	IndexManager  = index.IndexManager

	TapEntry        = index.TapEntry
	TapKey          = opts.TapKey
	TapTokens       = index.TapTokens
	IndexTapEntry   = index.TapEntry
	IndexSaveOption = index.SaveOption
)

var (
	pt = fmt.Printf
	is = errors.Is
	as = errors.As

	ce = e5.CheckWithStacktrace
	he = e5.Handle
	we = e5.WrapWithStacktrace

	Select        = index.Select
	Desc          = index.Desc
	Tap           = index.Tap
	Unmarshal     = index.Unmarshal
	TapPre        = index.TapPre
	Count         = index.Count
	Asc           = index.Asc
	NewEntry      = index.NewEntry
	MatchEntry    = index.MatchEntry
	MatchPreEntry = index.MatchPreEntry

	ErrKeyNotFound = store.ErrKeyNotFound
	ErrKeyNotMatch = store.ErrKeyNotMatch
)

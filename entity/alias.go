// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package entity

import (
	"errors"
	"fmt"

	"github.com/reusee/dscope"
	"github.com/reusee/e4"
	"github.com/reusee/june/index"
	"github.com/reusee/june/key"
	"github.com/reusee/june/opts"
	"github.com/reusee/june/store"
)

type (
	any = interface{}

	Key          = key.Key
	Hash         = key.Hash
	KeyPath      = key.KeyPath
	WriteResult  = store.WriteResult
	WriteOption  = store.WriteOption
	Order        = index.Order
	Store        = store.Store
	Scope        = dscope.Scope
	IndexEntry   = index.Entry
	NewHashState = key.NewHashState
	Limit        = index.Limit
	Index        = index.Index
	IndexManager = index.IndexManager

	TapKey          = opts.TapKey
	TapTokens       = index.TapTokens
	IndexTapEntry   = index.TapEntry
	IndexSaveOption = index.SaveOption
)

var (
	pt = fmt.Printf
	is = errors.Is
	as = errors.As

	ce = e4.CheckWithStacktrace
	he = e4.Handle
	we = e4.WrapWithStacktrace

	Select      = index.Select
	Desc        = index.Desc
	Call        = index.Call
	Tap         = index.Tap
	Count       = index.Count
	IndexEvSave = index.EvSave
	Asc         = index.Asc
	NewEntry    = index.NewEntry
	MatchEntry  = index.MatchEntry

	ErrKeyNotFound = store.ErrKeyNotFound
	ErrKeyNotMatch = store.ErrKeyNotMatch
)

// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package file

import (
	"errors"
	"fmt"

	"github.com/reusee/dscope"
	"github.com/reusee/e4"
	"github.com/reusee/june/entity"
	"github.com/reusee/june/filebase"
	"github.com/reusee/june/fsys"
	"github.com/reusee/june/index"
	"github.com/reusee/june/key"
	"github.com/reusee/june/store"
	"github.com/reusee/pp"
)

type (
	any = interface{}

	Key = key.Key

	Values = pp.Values[any]

	Scope = dscope.Scope
	Src   = filebase.Src
	Sink  = filebase.Sink

	Limit = index.Limit

	Store = store.Store
)

var (
	pt = fmt.Printf
	is = errors.Is
	ce = e4.CheckWithStacktrace
	he = e4.Handle

	PathSeparator = fsys.PathSeparator

	Get           = filebase.Get
	Copy          = pp.Copy[any, Src, Sink]
	Seq           = pp.Seq[any, Src]
	CollectValues = pp.CollectValues[any, Sink]
	IterValues    = pp.IterValues[any, Src]

	Select     = index.Select
	MatchEntry = index.MatchEntry
	Count      = index.Count
	Tap        = index.Tap
	Desc       = index.Desc
	Asc        = index.Asc

	IdxType    = entity.IdxType
	NSSummary  = entity.NSSummary
	NSEntity   = entity.NSEntity
	IdxReferTo = entity.IdxReferTo
)

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

	Scope         = dscope.Scope
	Src           = filebase.Src
	Sink          = filebase.Sink
	IterItem      = filebase.IterItem
	FileLike      = filebase.FileLike
	FileInfo      = filebase.FileInfo
	FileInfoThunk = filebase.FileInfoThunk
	PackThunk     = filebase.PackThunk
	Virtual       = filebase.Virtual
	ZipItem       = filebase.ZipItem

	Values = pp.Values[IterItem]

	Limit = index.Limit

	Store = store.Store
)

var (
	pt = fmt.Printf
	is = errors.Is
	ce = e4.CheckWithStacktrace
	he = e4.Handle
	we = e4.WrapWithStacktrace

	PathSeparator = fsys.PathSeparator

	Copy          = pp.Copy[IterItem, Src, Sink]
	Discard       = pp.Discard[IterItem, Sink]
	Tee           = pp.Tee[IterItem, Src, Sink]
	Seq           = pp.Seq[IterItem, Src]
	CollectValues = pp.CollectValues[IterItem, Sink]
	IterValues    = pp.IterValues[IterItem, Src]
	TapSink       = pp.Tap[IterItem, Sink]

	Get = filebase.Get

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

// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package file

import (
	"errors"
	"fmt"

	"github.com/reusee/dscope"
	"github.com/reusee/e4"
	"github.com/reusee/ling/v2/entity"
	"github.com/reusee/ling/v2/fsys"
	"github.com/reusee/ling/v2/index"
	"github.com/reusee/ling/v2/key"
	"github.com/reusee/ling/v2/store"
	"github.com/reusee/pp"
)

type (
	any = interface{}

	Key = key.Key

	Scope = dscope.Scope
	Src   = pp.Src
	Sink  = pp.Sink

	Limit = index.Limit

	Store = store.Store
)

var (
	pt = fmt.Printf
	is = errors.Is
	ce = e4.Check
	he = e4.Handle

	Copy = pp.Copy
	Seq  = pp.Seq

	PathSeparator = fsys.PathSeparator

	Select     = index.Select
	MatchEntry = index.MatchEntry
	Count      = index.Count
	Call       = index.Call
	Tap        = index.Tap
	Desc       = index.Desc
	Asc        = index.Asc

	IdxType    = entity.IdxType
	NSSummary  = entity.NSSummary
	NSEntity   = entity.NSEntity
	IdxReferTo = entity.IdxReferTo
)

// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package entity

import (
	"sync"

	"github.com/reusee/june/store"
)

// OnMark is called when key is marked, may call more than once for the same key
type TapMarkKey func(Key)

func (_ TapMarkKey) IsGCOption() {}

type TapReachableObjects func(*sync.Map)

func (_ TapReachableObjects) IsGCOption() {}

type TapIterKey func(Key)

func (_ TapIterKey) IsGCOption() {}

type TapDeadObjects func([]DeadObject)

func (_ TapDeadObjects) IsGCOption() {}

type TapSweepDeadObject func(DeadObject)

func (_ TapSweepDeadObject) IsGCOption() {}

type TapSummary func(*Summary)

func (_ TapSummary) IsSaveOption() {}

func (_ TapSummary) IsSyncOption() {}

type TapDeleteIndex func(IndexEntry)

func (_ TapDeleteIndex) IsDeleteIndexOption() {}

func (_ TapDeleteIndex) IsCleanIndexOption() {}

func (_ TapDeleteIndex) IsIndexGCOption() {}

type TapBadSummary func(*Summary)

func (_ TapBadSummary) IsSyncOption() {}

type TapWriteResult = store.TapWriteResult

type Parallel int

func (_ Parallel) IsPushOption() {}

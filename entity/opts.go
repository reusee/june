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

func (TapMarkKey) IsGCOption() {}

type TapReachableObjects func(*sync.Map)

func (TapReachableObjects) IsGCOption() {}

type TapIterKey func(Key)

func (TapIterKey) IsGCOption() {}

type TapDeadObjects func([]DeadObject)

func (TapDeadObjects) IsGCOption() {}

type TapSweepDeadObject func(DeadObject)

func (TapSweepDeadObject) IsGCOption() {}

type TapSummary func(*Summary)

func (TapSummary) IsSaveOption() {}

func (TapSummary) IsSyncOption() {}

type TapDeleteIndex func(IndexEntry)

func (TapDeleteIndex) IsDeleteIndexOption() {}

func (TapDeleteIndex) IsCleanIndexOption() {}

func (TapDeleteIndex) IsIndexGCOption() {}

type TapBadSummary func(*Summary)

func (TapBadSummary) IsSyncOption() {}

type TapWriteResult = store.TapWriteResult

type Parallel int

func (Parallel) IsPushOption() {}

type IgnoreSummary func(Summary) bool

func (IgnoreSummary) IsPushOption() {}

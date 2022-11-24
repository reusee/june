// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package fsys

import "time"

type TapUpdatePaths func([]string)

func (TapUpdatePaths) IsWatchOption() {}

type SingleDevice bool

func (SingleDevice) IsWatchOption() {}

type OnInitDone func()

func (OnInitDone) IsWatchOption() {}

type OnUpdatedSpec struct {
	MaxDelay time.Duration
	Func     func()
}

func (OnUpdatedSpec) IsWatchOption() {}

func OnUpdated(maxDelay time.Duration, fn func()) OnUpdatedSpec {
	return OnUpdatedSpec{
		MaxDelay: maxDelay,
		Func:     fn,
	}
}

type TrackFiles bool

func (TrackFiles) IsWatchOption() {}

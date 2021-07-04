// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package fsys

import "time"

type TapUpdatePaths func([]string)

func (_ TapUpdatePaths) IsWatchOption() {}

type SingleDevice bool

func (_ SingleDevice) IsWatchOption() {}

type OnInitDone func()

func (_ OnInitDone) IsWatchOption() {}

type OnUpdatedSpec struct {
	MaxDelay time.Duration
	Func     func()
}

func (_ OnUpdatedSpec) IsWatchOption() {}

func OnUpdated(maxDelay time.Duration, fn func()) OnUpdatedSpec {
	return OnUpdatedSpec{
		MaxDelay: maxDelay,
		Func:     fn,
	}
}

type TrackFiles bool

func (_ TrackFiles) IsWatchOption() {}

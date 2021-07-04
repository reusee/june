// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package clock

import (
	"time"
)

type Now func() time.Time

func (_ Def) Now() Now {
	return func() time.Time {
		return time.Now()
	}
}

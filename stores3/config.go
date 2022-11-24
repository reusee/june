// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package stores3

import (
	"time"
)

type Timeout time.Duration

func (Def) Timeout() Timeout {
	return Timeout(time.Minute * 8)
}

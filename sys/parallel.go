// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package sys

import "runtime"

type Parallel int

func (_ Def) Parallel() Parallel {
	parallel := runtime.NumCPU()
	if parallel < 2 {
		parallel = 2
	}
	return Parallel(parallel)
}

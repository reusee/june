// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package file

import (
	"github.com/reusee/june/sys"
)

func (_ Def) Configs(
	testing sys.Testing,
) (
	packThreshold PackThreshold,
	smallFileThreshold SmallFileThreshold,
) {

	// defaults
	packThreshold = 128
	smallFileThreshold = 4 * 1024
	if testing {
		packThreshold = 4
		smallFileThreshold = 128
	}

	return
}

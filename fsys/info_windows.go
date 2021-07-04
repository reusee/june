// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package fsys

import (
	"os"
)

func GetDevice(stat os.FileInfo) uint64 {
	// assume all files are in the same device
	return 0
}

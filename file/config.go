// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package file

import (
	"github.com/reusee/ling/v2/config"
	"github.com/reusee/ling/v2/sys"
)

func (_ Def) Configs(
	getConfig config.GetConfig,
	testing sys.Testing,
) (
	packThreshold PackThreshold,
	smallFileThreshold SmallFileThreshold,
) {

	var data struct {
		File struct {
			PackThreshold      PackThreshold
			SmallFileThreshold SmallFileThreshold
		}
	}

	// defaults
	data.File.PackThreshold = 128
	data.File.SmallFileThreshold = 4 * 1024
	if testing {
		data.File.PackThreshold = 4
		data.File.SmallFileThreshold = 128
	}

	// get
	err := getConfig(&data)
	ce(err)

	packThreshold = data.File.PackThreshold
	smallFileThreshold = data.File.SmallFileThreshold

	return
}

// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storepebble

import (
	"github.com/reusee/ling/v2/config"
)

type CacheSize int64

func (_ Def) CacheSize(
	getConfig config.GetConfig,
) (
	cacheSize CacheSize,
) {

	cacheSize = 32 * 1024 * 1024

	var config struct {
		Pebble struct {
			CacheSize int64
		}
	}
	err := getConfig(&config)
	ce(err)
	if config.Pebble.CacheSize != 0 {
		cacheSize = CacheSize(config.Pebble.CacheSize)
	}

	return
}

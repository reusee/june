// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storepebble

type CacheSize int64

func (Def) CacheSize() (
	cacheSize CacheSize,
) {

	cacheSize = 32 * 1024 * 1024

	return
}

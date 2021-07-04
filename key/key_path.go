// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package key

type KeyPath []Key

func (k KeyPath) Same(b KeyPath) bool {
	if len(k) == 0 && len(b) == 0 {
		return true
	}
	if len(k) == 0 || len(b) == 0 {
		return false
	}
	return k[len(k)-1] == b[len(b)-1]
}

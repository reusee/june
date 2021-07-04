// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package fsys

import (
	"os"
)

type EnsureDir func(path string) error

func (_ Def) EnsureDir() EnsureDir {
	return func(path string) (err error) {
		defer he(&err)
		_, err = os.Stat(path)
		if is(err, os.ErrNotExist) {
			err := os.MkdirAll(path, 0777)
			ce(err)
		} else {
			ce(err)
		}
		return nil
	}
}

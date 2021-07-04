// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package fsys

import (
	"os"
	"path/filepath"
)

func RealPath(path string) (ret string, err error) {
	defer he(&err)
	abs, err := filepath.Abs(path)
	ce(err)
	r, err := filepath.EvalSymlinks(abs)
	if is(err, os.ErrNotExist) {
		return abs, nil
	}
	ce(err)
	abs, err = filepath.Abs(r)
	ce(err)
	return abs, nil
}

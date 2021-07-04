// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package virtualfs

import (
	"io/fs"
	"testing"
)

func TestProjectedFS(
	t *testing.T,
	testFS TestFS,
	newFS NewProjectedFS,
) {
	testFS(t, func(
		rootFS fs.FS,
		dir string,
		fn func(),
	) {
		cl, err := newFS(rootFS, dir)
		ce(err)
		defer cl()
		fn()
	})
}

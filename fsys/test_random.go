// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package fsys

import (
	"testing"

	"github.com/reusee/e5"
)

func TestShuffleDir(
	t *testing.T,
	shuffle ShuffleDir,
) {
	defer he(nil, e5.TestingFatal(t))
	dir := t.TempDir()
	for i := 0; i < 1024; i++ {
		op, path1, path2, err := shuffle(dir)
		ce(err)
		if op == "" {
			t.Fatal()
		}
		if path1 == "" {
			t.Fatal()
		}
		_ = path2
	}
}

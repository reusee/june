// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

// +build !step1 !step2 windows

package june

import (
	"github.com/reusee/june/virtualfs"

	"testing"
)

func Test_virtualfs_TestProjectedFS(t *testing.T) {
	t.Parallel()
	runTest(t, virtualfs.TestProjectedFS)
}

// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

//go:build !step1 || !step2 || windows
// +build !step1 !step2 windows

package ling

import (
	"github.com/reusee/ling/v2/virtualfs"

	"testing"
)

func Test_virtualfs_TestProjectedFS(t *testing.T) {
	t.Parallel()
	runTest(t, virtualfs.TestProjectedFS)
}

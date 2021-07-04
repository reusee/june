// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

//go:build linux || darwin
// +build linux darwin

package fsys

import (
	"os"
	"syscall"
)

func GetDevice(stat os.FileInfo) uint64 {
	return uint64(stat.Sys().(*syscall.Stat_t).Dev)
}

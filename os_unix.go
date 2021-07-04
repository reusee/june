// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

//go:build unix
// +build unix

package ling

import (
	"syscall"
)

func init() {

	// set limits
	var rLimit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		return
	}
	rLimit.Cur = rLimit.Max
	syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		return
	}

}

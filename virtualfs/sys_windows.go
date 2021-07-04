// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package virtualfs

import (
	"fmt"

	"golang.org/x/sys/windows"
)

type WindowsSyscallError struct {
	R1    uintptr
	R2    uintptr
	Errno windows.Errno
}

func (w *WindowsSyscallError) Error() string {
	return fmt.Sprintf("syscall error: %x %x %x", w.R1, w.R2, w.Errno)
}

func winErr(r1 uintptr, r2 uintptr, err error) error {
	if r1 == 0 {
		return nil
	}
	return &WindowsSyscallError{
		R1:    r1,
		R2:    r2,
		Errno: err.(windows.Errno),
	}
}

func hresult(x uintptr) uintptr {
	if x <= 0 {
		return x
	}
	return ((x) & 0x0000FFFF) | (7 << 16) | 0x80000000
}

func hresultNT(x uintptr) uintptr {
	return x | 0x10000000
}

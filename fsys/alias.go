// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package fsys

import (
	"errors"
	"fmt"

	"github.com/reusee/e5"
)

var (
	pt = fmt.Printf
	is = errors.Is
	ce = e5.CheckWithStacktrace
	he = e5.Handle
)

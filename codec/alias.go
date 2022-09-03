// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package codec

import (
	"fmt"

	"github.com/reusee/e5"
)

var (
	ce = e5.CheckWithStacktrace
	he = e5.Handle
	we = e5.WrapWithStacktrace

	pt = fmt.Printf
)

type (
	any = interface{}
)

// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package key

import (
	"fmt"

	"github.com/reusee/e5"
)

var (
	ce = e5.Check.With(e5.WrapStacktrace)
	he = e5.Handle
	we = e5.Wrap.With(e5.WrapStacktrace)
	pt = fmt.Printf
)

type (
	any = interface{}
)

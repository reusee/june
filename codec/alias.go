// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package codec

import (
	"fmt"

	"github.com/reusee/e4"
)

var (
	ce = e4.Check
	he = e4.Handle
	we = e4.Wrap

	pt = fmt.Printf
)

type (
	any = interface{}
)

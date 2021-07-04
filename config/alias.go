// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package config

import (
	"fmt"

	"github.com/reusee/dscope"
	"github.com/reusee/e4"
)

var (
	catch = e4.Handle
	pt    = fmt.Printf
	ce    = e4.Check
	he    = e4.Handle
)

type (
	any = interface{}

	Scope = dscope.Scope
)

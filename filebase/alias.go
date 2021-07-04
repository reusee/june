// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package filebase

import (
	"errors"
	"fmt"

	"github.com/reusee/dscope"
	"github.com/reusee/e4"
	"github.com/reusee/june/key"
)

type (
	any = interface{}

	Key = key.Key

	Scope = dscope.Scope
)

var (
	ce, he = e4.Check, e4.Handle
	we     = e4.Wrap
	is     = errors.Is
	pt     = fmt.Printf
)

// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package filebase

import (
	"errors"
	"fmt"

	"github.com/reusee/dscope"
	"github.com/reusee/e5"
	"github.com/reusee/june/key"
)

type (
	any = interface{}

	Key = key.Key

	Scope = dscope.Scope
)

var (
	ce, he = e5.CheckWithStacktrace, e5.Handle
	we     = e5.WrapWithStacktrace
	is     = errors.Is
	pt     = fmt.Printf
)

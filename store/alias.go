// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package store

import (
	"errors"

	"github.com/reusee/e5"
	"github.com/reusee/june/key"
)

type (
	any  = interface{}
	Hash = key.Hash
)

var (
	as = errors.As
	ce = e5.CheckWithStacktrace
	he = e5.Handle
	we = e5.WrapWithStacktrace
)

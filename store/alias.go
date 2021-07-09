// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package store

import (
	"errors"

	"github.com/reusee/e4"
	"github.com/reusee/june/key"
)

type (
	any  = interface{}
	Hash = key.Hash
)

var (
	as = errors.As
	ce = e4.CheckWithStacktrace
	he = e4.Handle
	we = e4.Wrap
)

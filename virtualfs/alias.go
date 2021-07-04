// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package virtualfs

import (
	"errors"
	"fmt"

	"github.com/reusee/e4"
	"github.com/reusee/june/file"
	"github.com/reusee/june/key"
)

var (
	ce, he = e4.Check, e4.Handle
	as     = errors.As
	pt     = fmt.Printf
	is     = errors.Is
)

type (
	any      = interface{}
	Key      = key.Key
	ZipItem  = file.ZipItem
	FileInfo = file.FileInfo
)

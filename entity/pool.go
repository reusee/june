// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package entity

import (
	"bytes"

	"github.com/reusee/pr"
)

var bytesPool = pr.NewPool(32, func() any {
	return new(bytes.Buffer)
})

// Copyright 2022 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package juneerr

import "github.com/reusee/e5"

var (
	Handle = e5.Handle
	Wrap   = e5.Wrap.With(e5.WrapStacktrace)
	Check  = e5.Check.With(e5.WrapStacktrace)
)

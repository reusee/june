// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package index

import (
	"github.com/reusee/sb"
)

func IterStreams(streams []sb.Proc) ProcSrc {
	return func() (*sb.Proc, ProcSrc, error) {
		if len(streams) == 0 {
			return nil, nil, nil
		}
		return &streams[0], IterStreams(streams[1:]), nil
	}
}

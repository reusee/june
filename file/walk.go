// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package file

import (
	"fmt"
)

type Walk func(
	fn WalkFunc,
) Sink

type WalkFunc = func(string, FileLike) error

func (_ Def) Walk() (walk Walk) {

	walk = func(fn WalkFunc) Sink {
		var sink Sink
		sink = func(value *IterItem) (_ Sink, err error) {
			defer he(&err)
			if value == nil {
				return nil, nil
			}

			if value.FileInfo != nil {
				err := fn(value.FileInfo.Path, value.FileInfo.FileLike)
				ce(err)

			} else if value.FileInfoThunk != nil {
				err := fn(value.FileInfoThunk.FileInfo.Path, value.FileInfoThunk.FileInfo.FileLike)
				ce(err)
				value.FileInfoThunk.Expand(true)

			} else if value.PackThunk != nil {
				value.PackThunk.Expand(true)

			} else {
				ce(fmt.Errorf("unknown type %T", value))
			}

			return sink, nil
		}
		return sink
	}

	return
}

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
		sink = func(value *any) (_ Sink, err error) {
			defer he(&err)
			if value == nil {
				return nil, nil
			}
			switch value := (*value).(type) {

			case FileInfo:
				err := fn(value.Path, value.FileLike)
				ce(err)

			case FileInfoThunk:
				err := fn(value.FileInfo.Path, value.FileInfo.FileLike)
				ce(err)
				value.Expand(true)

			case PackThunk:
				value.Expand(true)

			default:
				panic(fmt.Errorf("unknown type %T", value))

			}
			return sink, nil
		}
		return sink
	}

	return
}

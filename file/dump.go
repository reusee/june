// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package file

func dump() Sink {
	var sink Sink
	sink = func(v *IterItem) (Sink, error) {
		if v == nil {
			return nil, nil
		}
		if v.FileInfoThunk != nil {
			v.FileInfoThunk.Expand(true)
			v.FileInfo = &v.FileInfoThunk.FileInfo
			v.FileInfoThunk = nil
		}
		return sink, nil
	}
	return sink
}

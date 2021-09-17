// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package file

func dump() Sink {
	var sink Sink
	sink = func(v *any) (Sink, error) {
		if v == nil {
			return nil, nil
		}
		if t, ok := (*v).(FileInfoThunk); ok {
			t.Expand(true)
			i := any(t.FileInfo)
			v = &i
		}
		return sink, nil
	}
	return sink
}

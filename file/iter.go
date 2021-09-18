// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package file

func invalid() (*IterItem, Src, error) {
	panic("invalid")
}

type Ignore func(
	path string, // relative path
	fileLike FileLike,
) bool

func (_ Def) DefaultIgnore() Ignore {
	return func(string, FileLike) bool {
		return false
	}
}

func ExpandAll(src Src) Src {
	var ret Src
	ret = func() (*IterItem, Src, error) {
		v, err := Get(&src)
		if err != nil {
			return nil, nil, err
		}
		if v == nil {
			return nil, nil, nil
		}
		if v.FileInfoThunk != nil {
			v.FileInfoThunk.Expand(true)
			v.FileInfo = &v.FileInfoThunk.FileInfo
			v.FileInfoThunk = nil
		} else if v.PackThunk != nil {
			v.PackThunk.Expand(true)
			v.PackThunk = nil
		}
		return v, ret, nil
	}
	return ret
}

func ExpandFileInfoThunk(src Src) Src {
	var ret Src
	ret = func() (*IterItem, Src, error) {
		v, err := Get(&src)
		if err != nil {
			return nil, nil, err
		}
		if v == nil {
			return nil, nil, nil
		}
		if v.FileInfoThunk != nil {
			v.FileInfoThunk.Expand(true)
			v.FileInfo = &v.FileInfoThunk.FileInfo
			v.FileInfoThunk = nil
		}
		return v, ret, nil
	}
	return ret
}

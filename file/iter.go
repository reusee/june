// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package file

type FileInfoThunk struct {
	FileInfo FileInfo
	Expand   func(bool)
	Path     string // relative path of file
}

type PackThunk struct {
	Expand func(bool)
	Path   string
	Pack   // relative path of dir
}

func invalid() (any, Src, error) {
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
	ret = func() (any, Src, error) {
		v, err := src.Next()
		if err != nil {
			return nil, nil, err
		}
		if v == nil {
			return nil, nil, nil
		}
		if t, ok := v.(FileInfoThunk); ok {
			v = t.FileInfo
			t.Expand(true)
		} else if t, ok := v.(PackThunk); ok {
			v = nil
			t.Expand(true)
		}
		return v, ret, nil
	}
	return ret
}

func ExpandFileInfoThunk(src Src) Src {
	var ret Src
	ret = func() (any, Src, error) {
		v, err := src.Next()
		if err != nil {
			return nil, nil, err
		}
		if v == nil {
			return nil, nil, nil
		}
		if t, ok := v.(FileInfoThunk); ok {
			v = t.FileInfo
			t.Expand(true)
		}
		return v, ret, nil
	}
	return ret
}

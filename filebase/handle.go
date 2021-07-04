// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package filebase

import (
	"io"
	"io/fs"

	"github.com/reusee/pp"
)

type Handle struct {
	io.ReadSeeker
	file     *File
	subsIter pp.Src
}

var _ fs.File = new(Handle)

var _ fs.ReadDirFile = new(Handle)

func (h *Handle) Stat() (fs.FileInfo, error) {
	return Info{
		file: h.file,
	}, nil
}

func (h *Handle) ReadDir(n int) (ret []fs.DirEntry, err error) {
	defer he(&err)

	if n <= 0 {
		for {
			v, err := h.subsIter.Next()
			if err != nil {
				return ret, err
			}
			if v == nil {
				break
			}
			file := v.(*File)
			ret = append(ret, DirEntry{
				file: file,
			})
		}

	} else {
		for n > 0 {
			v, err := h.subsIter.Next()
			if err != nil {
				return ret, err
			}
			if v == nil {
				return ret, io.EOF
			}
			file := v.(*File)
			ret = append(ret, DirEntry{
				file: file,
			})
			n--
		}
	}

	return
}

func (h *Handle) Close() error {
	return nil
}

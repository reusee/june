// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package filebase

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"strings"
	"sync"

	"github.com/reusee/pp"
)

var ErrNotDir = errors.New("not a dir")

type NewFileFS func(file *File) (fs.FS, error)

func (_ Def) NewFileFS(
	iterSubs IterSubs,
	newContentReadre NewContentReader,
	findFileInSubs FindFileInSubs,
) NewFileFS {
	return func(file *File) (fs.FS, error) {
		if !file.IsDir {
			return nil, we(ErrNotDir)
		}
		return &FS{
			file: file,

			iterSubs:         iterSubs,
			newContentReader: newContentReadre,
			findFileInSubs:   findFileInSubs,
		}, nil
	}
}

type FS struct {
	file *File

	cache sync.Map

	iterSubs         IterSubs
	newContentReader NewContentReader
	findFileInSubs   FindFileInSubs
}

var _ fs.FS = new(FS)

func (f *FS) Open(name string) (_ fs.File, err error) {
	defer he(&err)

	file, err := f.open(name)
	ce(err)

	var r io.ReadSeeker
	var iter pp.Src

	if file.IsDir {
		iter = f.iterSubs(file.Subs, nil)
	} else {
		if len(file.Contents) > 0 {
			r = f.newContentReader(file.Contents, file.ChunkLengths)
		} else {
			r = bytes.NewReader(file.ContentBytes)
		}
	}

	return &Handle{
		ReadSeeker: r,
		file:       file,
		subsIter:   iter,
	}, nil
}

func (f *FS) open(name string) (file *File, err error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{
			Op:   "open",
			Path: name,
			Err:  fs.ErrInvalid,
		}
	}

	v, ok := f.cache.Load(name)
	if ok {
		return v.(*File), nil
	}

	defer func() {
		if err == nil {
			f.cache.Store(name, file)
		}
	}()

	if name == "." {
		return f.file, nil
	}

	file, err = f.findFileInSubs(
		f.file.Subs,
		strings.Split(name, "/"),
	)
	if err != nil {
		return nil, &fs.PathError{
			Op:   "open",
			Path: name,
			Err:  err,
		}
	}
	if file == nil {
		return nil, &fs.PathError{
			Op:   "open",
			Path: name,
			Err:  fs.ErrNotExist,
		}
	}

	return file, nil
}

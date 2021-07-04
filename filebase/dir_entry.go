// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package filebase

import "io/fs"

type DirEntry struct {
	file *File
}

var _ fs.DirEntry = DirEntry{}

func (e DirEntry) Info() (fs.FileInfo, error) {
	return Info(e), nil
}

func (e DirEntry) IsDir() bool {
	return e.file.IsDir
}

func (e DirEntry) Type() fs.FileMode {
	return e.file.Mode.Type()
}

func (e DirEntry) Name() string {
	return e.file.Name
}

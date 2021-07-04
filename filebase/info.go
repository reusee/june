// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package filebase

import (
	"io/fs"
	"time"
)

type Info struct {
	file *File
}

var _ fs.FileInfo = Info{}

func (i Info) IsDir() bool {
	return i.file.IsDir
}

func (i Info) ModTime() time.Time {
	return i.file.ModTime
}

func (i Info) Mode() fs.FileMode {
	return i.file.Mode
}

func (i Info) Name() string {
	return i.file.Name
}

func (i Info) Size() int64 {
	return i.file.Size
}

func (i Info) Sys() any {
	return i.file
}

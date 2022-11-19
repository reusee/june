// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package file

import (
	"bytes"
	"context"
	"io"
	"os"
	"time"

	"github.com/reusee/e5"
	"github.com/reusee/june/fsys"
)

type DiskFile struct {
	info         os.FileInfo
	Path         string
	content      []byte
	contentReady bool
}

var _ FileLike = DiskFile{}

func (d DiskFile) GetIsDir(_ Scope) bool {
	return d.info.IsDir()
}

func (d DiskFile) GetName(_ Scope) string {
	return d.info.Name()
}

func (d DiskFile) GetSize(_ Scope) int64 {
	if d.contentReady {
		return int64(len(d.content))
	}
	if d.info.IsDir() {
		return 0
	}
	return d.info.Size()
}

func (d DiskFile) GetMode(_ Scope) os.FileMode {
	return d.info.Mode()
}

func (d DiskFile) GetModTime(_ Scope) time.Time {
	return d.info.ModTime()
}

func (d DiskFile) GetDevice(_ Scope) uint64 {
	return fsys.GetDevice(d.info)
}

func (d DiskFile) WithReader(ctx context.Context, scope Scope, fn func(io.Reader) error) (err error) {
	defer he(&err, e5.Info("read %s", d.Path))
	var r io.Reader
	if d.contentReady {
		r = bytes.NewReader(d.content)
	} else {
		f, err := os.Open(d.Path)
		ce(err)
		defer f.Close()
		r = f
	}
	err = fn(r)
	ce(err)
	return nil
}

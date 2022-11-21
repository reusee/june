// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package file

import (
	"bytes"
	"io"
	"os"
	"time"
)

type Virtual struct {
	ModTime time.Time
	Name    string
	Subs    []Virtual
	Content []byte
	Size    int64
	Mode    os.FileMode
	IsDir   bool
}

var _ FileLike = new(Virtual)

func (v Virtual) GetIsDir(_ Scope) bool {
	return v.IsDir
}

func (v Virtual) GetName(_ Scope) string {
	return v.Name
}

func (v Virtual) GetSize(_ Scope) int64 {
	return v.Size
}

func (v Virtual) GetMode(_ Scope) os.FileMode {
	return v.Mode
}

func (v Virtual) GetModTime(_ Scope) time.Time {
	return v.ModTime
}

func (v Virtual) GetDevice(_ Scope) uint64 {
	return 0
}

func (v Virtual) WithReader(scope Scope, fn func(io.Reader) error) error {
	return fn(bytes.NewReader(v.Content))
}

// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package file

import (
	"io"
	"os"
	"time"

	"github.com/reusee/e5"
)

type FileLike interface {
	// basic infos
	GetIsDir(Scope) bool
	GetName(Scope) string
	GetSize(Scope) int64
	GetMode(Scope) os.FileMode
	GetModTime(Scope) time.Time
	GetDevice(Scope) uint64

	// content
	WithReader(
		Scope,
		func(io.Reader) error,
	) error
}

func writeFileLikeToDisk(scope Scope, value FileLike, path string) (err error) {
	defer he(&err, e5.Info("write to %s", path))
	f, err := os.OpenFile(
		path,
		os.O_RDWR|os.O_CREATE,
		value.GetMode(scope),
	)
	ce(err)
	defer f.Close()
	if err := value.WithReader(scope, func(r io.Reader) error {
		_, err := io.Copy(f, r)
		return err
	}); err != nil {
		return err
	}
	err = f.Close()
	ce(err)
	modTime := value.GetModTime(scope)
	err = os.Chtimes(path, modTime, modTime)
	ce(err)
	return nil
}

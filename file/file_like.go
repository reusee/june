// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package file

import (
	"io"
	"os"

	"github.com/reusee/e4"
)

func writeFileLikeToDisk(scope Scope, value FileLike, path string) (err error) {
	defer he(&err, e4.NewInfo("write to %s", path))
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

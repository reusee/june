// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package file

import (
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/reusee/e4"
	"github.com/reusee/ling/v2/filebase"
	"github.com/reusee/pp"
)

func TestFileFS(
	t *testing.T,
	build Build,
	iterDisk IterDiskFile,
	newFileFS filebase.NewFileFS,
) {
	defer he(nil, e4.TestingFatal(t))

	var root File
	ce(pp.Copy(
		iterDisk("testdata", nil),
		build(&root, nil),
	))

	f, err := newFileFS(root.Subs[0].File)
	ce(err)
	ce(fstest.TestFS(f,
		".gitignore",
		"zip",
	))

	n := 0
	ce(fs.WalkDir(f, ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		n++
		return nil
	}))
	if n != 160 {
		t.Fatal()
	}

}

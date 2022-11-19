// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package file

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/reusee/e5"
	"github.com/reusee/june/filebase"
	"github.com/reusee/june/fsys"
	"github.com/reusee/pp"
)

func TestFileFS(
	t *testing.T,
	build Build,
	iterDisk IterDiskFile,
	newFileFS filebase.NewFileFS,
	shuffleDir fsys.ShuffleDir,
) {
	defer he(nil, e5.TestingFatal(t))
	ctx := context.Background()

	dir := t.TempDir()
	for i := 0; i < 64; i++ {
		_, _, _, err := shuffleDir(dir)
		ce(err)
	}
	var expectes []string
	var numFiles int
	ce(filepath.WalkDir(dir, func(path string, _ fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		numFiles++
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		rel = strings.ReplaceAll(rel, string([]rune{os.PathSeparator}), "/")
		expectes = append(expectes, rel)
		return nil
	}))

	var root File
	ce(pp.Copy(
		iterDisk(ctx, dir, nil),
		build(ctx, &root, nil),
	))

	f, err := newFileFS(ctx, root.Subs[0].File)
	ce(err)
	ce(fstest.TestFS(f,
		expectes...,
	))

	n := 0
	ce(fs.WalkDir(f, ".", func(_ string, _ fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		n++
		return nil
	}))
	if n != numFiles {
		t.Fatalf("got %d, expected %d", n, numFiles)
	}

}

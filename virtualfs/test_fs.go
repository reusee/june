// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package virtualfs

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/reusee/e4"
	"github.com/reusee/june/entity"
	"github.com/reusee/june/file"
	"github.com/reusee/june/filebase"
	"github.com/reusee/june/fsys"
	"github.com/reusee/pp"
)

type TestFS func(
	t *testing.T,
	with func(
		rootFS fs.FS,
		dir string,
		fn func(),
	),
)

func (_ Def) TesetFS(
	iterDisk file.IterDiskFile,
	build file.Build,
	equal file.Equal,
	zip file.Zip,
	save entity.SaveEntity,
	iterFile file.IterFile,
	newFileFS filebase.NewFileFS,
	shuffleDir fsys.ShuffleDir,
) TestFS {

	return func(
		t *testing.T,
		with func(
			rootFS fs.FS,
			dir string,
			fn func(),
		),
	) {
		defer he(nil, e4.TestingFatal(t))

		dataDir := t.TempDir()
		for i := 0; i < 64; i++ {
			shuffleDir(dataDir)
		}
		dataDirName := filepath.Base(dataDir)
		var numFiles int
		ce(filepath.WalkDir(dataDir, func(_ string, _ fs.DirEntry, err error) error {
			numFiles++
			return err
		}))

		root := new(filebase.File)
		ce(pp.Copy(
			iterDisk(dataDir, nil),
			build(root, nil),
		))
		file1 := root.Subs[0].File
		summary, err := save(file1)
		ce(err)
		key1 := summary.Key

		dir := t.TempDir()
		f, err := newFileFS(root)
		ce(err)

		with(f, dir, func() {

			_, err = os.Stat(filepath.Join(dir, "foo"))
			if !is(err, os.ErrNotExist) {
				t.Fatal()
			}

			names, err := filepath.Glob(filepath.Join(dir, "*"))
			ce(err)
			if len(names) != 1 {
				t.Fatal()
			}

			ce(filepath.WalkDir(dataDir, func(path string, entry fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				rel, err := filepath.Rel(dataDir, path)
				ce(err)
				rel = filepath.Join(dataDirName, rel)
				_, err = os.Stat(filepath.Join(dir, rel))
				ce(err)
				return nil
			}))

			var n int
			ce(filepath.Walk(dir, func(
				path string,
				info os.FileInfo,
				err error,
			) error {
				if err != nil {
					t.Fatal(err)
				}
				n++
				return nil
			}))
			if n != numFiles+1 {
				t.Fatalf("expected %d, got %d\n", numFiles, n)
			}

			var numFile2 int
			ce(filepath.WalkDir(dir, func(
				path string,
				entry fs.DirEntry,
				err error,
			) error {
				if err != nil {
					return err
				}
				numFile2++
				return nil
			}))
			if numFile2 != numFiles+1 {
				t.Fatal()
			}

			iter := zip(
				iterDisk(filepath.Join(dir, dataDirName), nil),
				iterDisk(dataDir, nil),
				nil,
			)
			for {
				v, err := file.Get(&iter)
				ce(err)
				if v == nil {
					break
				}
				item := v.ZipItem
				if item.A == nil && item.B != nil {
					t.Fatalf("B: %#v\n", item.B)
				} else if item.A != nil && item.B == nil {
					t.Fatalf("A: %#v\n", item.A)
				}
			}

			ok, err := equal(
				iterDisk(filepath.Join(dir, dataDirName), nil),
				iterDisk(dataDir, nil),
				func(a, b any, reason string) {
					pt("DIFF %s\n\t%#v\n\t%#v\n\n", reason, a, b)
				},
			)
			ce(err)
			if !ok {
				t.Fatal()
			}

			var root2 filebase.File
			ce(pp.Copy(
				iterDisk(filepath.Join(dir, dataDirName), nil),
				build(&root2, nil),
			))
			file2 := root2.Subs[0].File
			summary, err = save(file2)
			ce(err)
			key2 := summary.Key
			if key2 != key1 {
				ok, err := equal(
					iterFile(file1, nil),
					iterFile(file2, nil),
					func(a, b any, reason string) {
						pt("DIFF %s\n\t%#v\n\t%#v\n\n", reason, a, b)
					},
				)
				ce(err)
				if !ok {
					t.Fatal()
				}
				t.Fatal()
			}

		})

	}

}

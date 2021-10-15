// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package file

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/reusee/e4"
	"github.com/reusee/june/entity"
	"github.com/reusee/june/fsys"
	"github.com/reusee/june/index"
	"github.com/reusee/june/store"
	"github.com/reusee/pp"
)

func TestSave(
	t *testing.T,
	walk Walk,
	write WriteContents,
	store store.Store,
	fetch entity.Fetch,
	scope Scope,
) {
	defer he(nil, e4.TestingFatal(t))

	scope.Fork(
		func() Ignore {
			return func(path string, file FileLike) bool {
				// ignore .git dir
				if file.GetIsDir(scope) && strings.Contains(path, ".git") {
					return true
				}
				return false
			}
		},
	).Call(func(
		iterDiskFile IterDiskFile,
		iterFile IterFile,
		build Build,
		iterKey IterKey,
		equal Equal,
		fetch entity.Fetch,
		checkRef entity.CheckRef,
		store Store,
		gc entity.GC,
		rebuildIndex entity.RebuildIndex,
		index index.Index,
	) {

		c := 0
		numRead := 0

		// save
		file := new(File)
		err := Copy(
			iterDiskFile(".", nil, UseGitIgnore(false)),
			build(
				file, nil,
				TapBuildFile(func(_ FileInfo, _ *File) {
					c++
				}),
				TapReadFile(func(_ FileInfo) {
					numRead++
				}),
			),
		)
		ce(err)
		if c == 0 {
			t.Fatal()
		}
		if numRead == 0 {
			t.Fatal()
		}
		file = file.Subs[0].File

		// equal
		ok, err := equal(
			iterDiskFile(".", nil, UseGitIgnore(false)),
			iterFile(file, nil),
			func(a, b any, reason string) {
				pt("NOT EQUAL: %s\n\t%#v\n\t%#v\n\n", reason, a, b)
			},
		)
		ce(err)
		if !ok {
			t.Fatal()
		}

		var n int
		if err := Select(
			index,
			entity.MatchType(Subs{}),
			Count(&n),
		); err != nil {
			t.Fatal(err)
		}
		if n == 0 {
			t.Fatal()
		}

		abs, err := fsys.RealPath(".")
		ce(err)
		dir := filepath.Dir(abs)
		n = 0
		if err := Copy(
			iterFile(file, nil),
			walk(func(path string, file FileLike) (err error) {
				defer he(&err)
				n++

				// file content
				if !file.GetIsDir(scope) {
					buf := new(bytes.Buffer)
					if err := file.WithReader(scope, func(r io.Reader) error {
						_, err := io.Copy(buf, r)
						return err
					}); err != nil {
						return err
					}
					content, err := os.ReadFile(
						filepath.Join(dir, path),
					)
					ce(err)
					if !bytes.Equal(buf.Bytes(), content) {
						return fmt.Errorf(
							"content not match: %s",
							filepath.Join(path, file.GetName(scope)),
						)
					}
				}

				return nil
			}),
		); err != nil {
			t.Fatal(err)
		}

		// file numbers
		m := 0
		if err := filepath.WalkDir(".", func(path string, entry fs.DirEntry, e error) error {
			if path == ".git" || path == ".github" {
				return fs.SkipDir
			}
			m++
			if e != nil {
				return e
			}
			return nil
		}); err != nil {
			t.Fatal(err)
		}
		if n != m {
			t.Fatalf("%d on disk, %d saved\n", m, n)
		}

		var paths []string

		// iter
		file2 := new(File)
		err = Copy(
			pp.Tee(
				iterFile(file, nil),
			),
			build(
				file2, nil,
				TapBuildFile(func(info FileInfo, file *File) {
					paths = append(paths, info.Path)
				}),
			),
		)
		ce(err)
		if len(paths) == 0 {
			t.Fatal()
		}
		file2 = file2.Subs[0].File

		// equal
		ok, err = equal(
			iterFile(file, nil),
			iterFile(file2, nil),
			func(a, b any, reason string) {
				pt("DIFF %s\n\t%#v\n\t%#v\n\n", reason, a, b)
			},
		)
		ce(err)
		if !ok {
			t.Fatal()
		}
		ok, err = equal(
			iterDiskFile(".", nil, UseGitIgnore(false)),
			iterFile(file2, nil),
			func(a, b any, reason string) {
				pt("DIFF %s\n\t%#v\n\t%#v\n\n", reason, a, b)
			},
		)
		ce(err)
		if !ok {
			t.Fatal()
		}

		// size
		if file.GetTreeSize(scope) < 256*1024 { // least source tree size
			t.Fatalf("got %d", file.GetTreeSize(scope))
		}

	})

}

func TestSymlink(
	t *testing.T,
	build Build,
	iterDiskfile IterDiskFile,
) {
	defer he(nil, e4.TestingFatal(t))

	if runtime.GOOS == "windows" {
		t.Skip()
	}

	dir := t.TempDir()

	err := os.Symlink(dir, filepath.Join(dir, "foo"))
	ce(err)
	file := new(File)
	err = Copy(
		iterDiskfile(dir, nil),
		build(file, nil),
	)
	ce(err)
}

func TestPack(
	t *testing.T,
	scope Scope,
) {
	defer he(nil, e4.TestingFatal(t))

	scope.Fork(
		func() PackThreshold {
			return 2
		},
	).Call(func(
		build Build,
		iter IterVirtual,
		fetch entity.Fetch,
	) {

		var tap Sink
		tap = func(v any) (Sink, error) {
			if v == nil {
				return nil, nil
			}
			return tap, nil
		}
		file := new(File)
		err := Copy(
			pp.Tee(
				iter(Virtual{
					Name:  "root",
					IsDir: true,
					Subs: []Virtual{
						{
							Name: "a",
						},
						{
							Name: "b",
						},
						{
							Name: "c",
						},
						{
							Name: "d",
						},
					},
				}, nil),
				tap,
			),
			build(file, nil),
		)
		ce(err)

		file = file.Subs[0].File
		if len(file.Subs) != 1 {
			t.Fatal()
		}

		pack := file.Subs[0].Pack
		if pack.Min != "a" {
			t.Fatal()
		}
		if pack.Max != "d" {
			t.Fatal()
		}
		if pack.Height != 3 {
			t.Fatal()
		}

		var ss Subs
		err = fetch(pack.Key, &ss)
		ce(err)
		pack = ss[0].Pack
		if pack.Min != "a" {
			t.Fatal()
		}
		if pack.Max != "b" {
			t.Fatal()
		}
		if pack.Height != 2 {
			t.Fatal()
		}
		pack = ss[1].Pack
		if pack.Min != "c" {
			t.Fatal()
		}
		if pack.Max != "d" {
			t.Fatal()
		}
		if pack.Height != 2 {
			t.Fatal()
		}

	})
}

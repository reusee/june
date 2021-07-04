// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package fsys

import (
	crand "crypto/rand"
	"io"
	"math/rand"
	"os"
	"path/filepath"

	"github.com/reusee/e4"
)

type ShuffleDir func(string) (
	op string,
	path1 string,
	path2 string,
	err error,
)

func (_ Def) ShuffleDir() ShuffleDir {

	return func(root string) (
		op string,
		path1 string,
		path2 string,
		err error,
	) {
		defer he(&err)

		var paths []string
		var dirs []string
		var files []string

		var collect func(dir string) error
		collect = func(path string) (err error) {
			defer he(&err)
			stat, err := os.Lstat(path)
			ce(err)
			if !stat.IsDir() {
				files = append(files, path)
				return nil
			}
			dirs = append(dirs, path)
			f, err := os.Open(path)
			ce(err)
			defer f.Close()
			names, err := f.Readdirnames(-1)
			ce(err)
			for _, name := range names {
				paths = append(paths, filepath.Join(path, name))
				err := collect(filepath.Join(path, name))
				ce(err)
			}
			return nil
		}
		err = collect(root)
		ce(err)

		add := func() (err error) {
			defer he(&err)
			op = "add"
			defer he(&err)
			dir := dirs[rand.Intn(len(dirs))]
			for {
				if rand.Intn(2) == 0 {
					// file
					var f *os.File
					f, err = os.CreateTemp(dir, "*")
					ce(err)
					path1 = f.Name()
					_, err = io.CopyN(f, crand.Reader, int64(rand.Intn(128)))
					ce(err, e4.Close(f))
					ce(f.Close())
					return
				} else {
					// dir
					name, err := os.MkdirTemp(dir, "*")
					ce(err)
					path1 = name
					dir = name
				}
			}
		}

		remove := func() (err error) {
			defer he(&err)
			path := paths[rand.Intn(len(paths))]
			op = "remove"
			path1, err = RealPath(path)
			ce(err)
			ce(os.RemoveAll(path))
			return
		}

		replace := func() (err error) {
			defer he(&err)
			op = "replace"
			defer he(&err)
			path := files[rand.Intn(len(files))]
			path1 = path
			f, err := os.Create(path)
			ce(err)
			_, err = io.CopyN(f, crand.Reader, int64(rand.Intn(128)))
			ce(err, e4.Close(f))
			ce(f.Close())
			return nil
		}

		truncate := func() (err error) {
			defer he(&err)
			op = "truncate"
			defer he(&err)
			path := files[rand.Intn(len(files))]
			path1 = path
			stat, err := os.Stat(path)
			ce(err)
			size := int(stat.Size())
			err = os.Truncate(path, int64(rand.Intn(size+1)))
			ce(err)
			return nil
		}

		var processes []func() error
		if len(dirs) > 0 {
			processes = append(processes, add)
		}
		if len(paths) > 0 {
			processes = append(processes, remove)
		}
		if len(files) > 0 {
			processes = append(processes, replace)
			processes = append(processes, truncate)
		}

		err = (processes[rand.Intn(len(processes))])()
		ce(err)

		return
	}

}

// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package file

import (
	"math/rand"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/reusee/e5"
	"github.com/reusee/june/fsys"
	"github.com/reusee/june/storekv"
	"github.com/reusee/june/storemem"
	"github.com/reusee/pr2"
)

func TestUpdate(
	t *testing.T,
	newKV storekv.New,
	newMem storemem.New,
	scope Scope,
	shuffle fsys.ShuffleDir,
	wg *pr2.WaitGroup,
) {
	defer he(nil, e5.TestingFatal(t))

	store, err := newKV(wg, newMem(wg), "test")
	ce(err)

	scope.Fork(func() Store {
		return store
	}).Call(func(
		build Build,
		iterDisk IterDiskFile,
		update Update,
		equal Equal,
		watch fsys.Watch,
		iterFile IterFile,
		wg *pr2.WaitGroup,
	) {

		dir := t.TempDir()
		watcher, err := watch(wg, dir)
		ce(err)

		// build
		t0 := time.Now()
		file := new(File)
		var numFile int64
		err = Copy(
			iterDisk(wg, dir, nil),
			build(
				wg,
				file,
				nil,
				TapBuildFile(func(info FileInfo, _ *File) {
					atomic.AddInt64(&numFile, 1)
					if filepath.Base(dir) != info.Path {
						t.Fatal()
					}
				}),
			),
		)
		ce(err)
		if numFile != 1 {
			t.Fatal()
		}
		file = file.Subs[0].File

		// update, same dir
		file2 := new(File)
		t1 := time.Now()
		err = Copy(
			update(
				dir,
				iterFile(wg, file, nil),
				t0,
				iterDisk(wg, dir, nil),
				watcher,
			),
			build(wg, file2, nil),
		)
		ce(err)
		file2 = file2.Subs[0].File
		if ok, err := equal(
			iterFile(wg, file, nil),
			iterFile(wg, file2, nil),
			nil,
		); err != nil {
			t.Fatal(err)
		} else if !ok {
			t.Fatal()
		}

		// randomize test
		lastFile := file2
		lastTime := t1
		var numRead int64
		for i := 0; i < 32; i++ {
			// shuffle
			n := rand.Intn(8)
			for i := 0; i < n; i++ {
				_, _, _, err := shuffle(dir)
				ce(err)
			}

			// wait watcher, set to longer if fail
			time.Sleep(time.Millisecond * 200)

			// update
			file := new(File)
			t2 := time.Now()
			err = Copy(
				update(
					dir,
					iterFile(wg, lastFile, nil),
					lastTime,
					iterDisk(wg, dir, nil),
					watcher,
				),
				build(
					wg,
					file,
					nil,
					TapReadFile(func(_ FileInfo) {
						atomic.AddInt64(&numRead, 1)
					}),
				),
			)
			ce(err)
			file = file.Subs[0].File

			// verify
			ok, err := equal(
				iterDisk(wg, dir, nil),
				iterFile(wg, file, nil),
				func(a, b any, reason string) {
					pt("DIFF %s\n\t%#v\n\t%#v\n\n", reason, a, b)
				},
			)
			ce(err)
			if !ok {
				t.Fatal()
			}

			lastFile = file
			lastTime = t2
		}

		if numRead == 0 {
			t.Fatal()
		}

	})

}

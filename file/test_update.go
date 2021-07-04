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

	"github.com/reusee/e4"
	"github.com/reusee/june/clock"
	"github.com/reusee/june/entity"
	"github.com/reusee/june/fsys"
	"github.com/reusee/june/index"
	"github.com/reusee/june/store"
	"github.com/reusee/june/storekv"
	"github.com/reusee/june/storemem"
	"github.com/reusee/june/storepebble"
	"github.com/reusee/june/tx"
)

func TestUpdate(
	t *testing.T,
	newKV storekv.New,
	newMem storemem.New,
	scope Scope,
	shuffle fsys.ShuffleDir,
) {
	defer he(nil, e4.TestingFatal(t))

	store, err := newKV(newMem(), "test")
	ce(err)
	defer store.Close()

	scope.Sub(func() Store {
		return store
	}).Call(func(
		build Build,
		iterDisk IterDiskFile,
		update Update,
		iterKey IterKey,
		equal Equal,
		watch fsys.Watch,
		getTime clock.Now,
		iterFile IterFile,
	) {

		dir := t.TempDir()
		watcher, err := watch(dir)
		ce(err)
		defer watcher.Close()

		// build
		t0 := getTime()
		file := new(File)
		var numFile int64
		err = Copy(
			iterDisk(dir, nil),
			build(
				file,
				nil,
				TapBuildFile(func(info FileInfo, file *File) {
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
		t1 := getTime()
		err = Copy(
			update(
				dir,
				iterFile(file, nil),
				t0,
				iterDisk(dir, nil),
				watcher,
			),
			build(file2, nil),
		)
		ce(err)
		file2 = file2.Subs[0].File
		if ok, err := equal(
			iterFile(file, nil),
			iterFile(file2, nil),
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
			t2 := getTime()
			err = Copy(
				update(
					dir,
					iterFile(lastFile, nil),
					lastTime,
					iterDisk(dir, nil),
					watcher,
				),
				build(
					file,
					nil,
					TapReadFile(func(info FileInfo) {
						atomic.AddInt64(&numRead, 1)
					}),
				),
			)
			ce(err)
			file = file.Subs[0].File

			// verify
			ok, err := equal(
				iterDisk(dir, nil),
				iterFile(file, nil),
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

func TestFileWithTx(
	t *testing.T,
	newPeb storepebble.New,
	watch fsys.Watch,
	newKV storekv.New,
	scope Scope,
) {
	defer he(nil, e4.TestingFatal(t))

	dir := t.TempDir()
	watcher, err := watch(dir)
	ce(err)
	defer watcher.Close()

	peb, err := newPeb(storepebble.NewMemFS(), "")
	ce(err)
	defer peb.Close()

	scope.Sub(
		func() tx.KVToStore {
			return func(kv storekv.KV) (store.Store, error) {
				return newKV(kv, "foo")
			}
		},
		tx.UsePebbleTx,
		&peb,

		// build with tx
		func(
			tx tx.PebbleTx,
		) BuildWithTx {
			return func(fn any) {
				ce(tx(fn))
			}
		},
	).Call(func(
		tx tx.PebbleTx,
	) {

		var key Key
		ce(tx(func(
			iterDisk IterDiskFile,
			build Build,
			save entity.SaveEntity,
		) {
			file := new(File)
			ce(Copy(
				iterDisk(dir, nil),
				build(
					file,
					nil,
				),
			))
			file = file.Subs[0].File
			s, err := save(file)
			ce(err)
			key = s.Key
		}))

		ce(tx(func(
			selIndex index.SelectIndex,
		) {
			var n int

			ce(selIndex(
				MatchEntry(entity.IdxSummaryKey, key),
				Count(&n),
			))
			if n != 1 {
				t.Fatal()
			}

			ce(selIndex(
				entity.MatchType(&File{}),
				Count(&n),
			))
			if n != 1 {
				t.Fatalf("got %d\n", n)
			}

		}))
		_ = key

	})

}

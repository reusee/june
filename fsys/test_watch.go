// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package fsys

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/reusee/e4"
)

func TestWatch(
	t *testing.T,
	watch Watch,
	shuffle ShuffleDir,
) {
	defer he(nil, e4.TestingFatal(t))

	// temp file
	dir := t.TempDir()
	ce(os.Mkdir(filepath.Join(dir, "foo"), 0755))
	ce(os.Mkdir(filepath.Join(dir, "bar"), 0755))
	f, err := os.CreateTemp(filepath.Join(dir, "foo"), "*")
	ce(err)
	ce(f.Close())
	time.Sleep(time.Millisecond * 200)

	// watch
	var counter int64
	var updatedCounter int64
	initOK := make(chan struct{})
	w, err := watch(
		dir,
		TapUpdatePaths(func(paths []string) {
			atomic.AddInt64(&counter, int64(len(paths)))
		}),
		OnUpdated(time.Millisecond*10, func() {
			atomic.AddInt64(&updatedCounter, 1)
		}),
		OnInitDone(func() {
			close(initOK)
		}),
		TrackFiles(true),
	)
	ce(err)
	<-initOK
	<-w.Ready

	stat, err := os.Stat(f.Name())
	ce(err)
	if w.root.ModTime.Before(stat.ModTime()) {
		t.Fatal()
	}

	shouldChanged := func(path string, ti time.Time) {
		t.Helper()
		notChanged, err := w.PathNotChanged(path, ti)
		ce(err)
		if notChanged {
			t.Fatal()
		}
	}
	shouldNotChanged := func(path string, ti time.Time) {
		t.Helper()
		notChanged, err := w.PathNotChanged(path, ti)
		ce(err)
		if !notChanged {
			t.Fatal()
		}
	}

	// change foo
	t0 := time.Now()
	f, err = os.CreateTemp(filepath.Join(dir, "foo"), "*")
	ce(err)
	ce(f.Close())
	w.waitChange(f.Name(), t0, time.Second*5)
	shouldChanged(dir, t0)
	shouldChanged(filepath.Join(dir, "foo"), t0)
	shouldNotChanged(filepath.Join(dir, "bar"), t0)

	// change bar
	t0 = time.Now()
	f, err = os.CreateTemp(filepath.Join(dir, "bar"), "*")
	ce(err)
	ce(f.Sync())
	ce(f.Close())
	w.waitChange(f.Name(), t0, time.Second*5)
	shouldChanged(dir, t0)
	shouldNotChanged(filepath.Join(dir, "foo"), t0)
	shouldChanged(filepath.Join(dir, "bar"), t0)

	// no change
	t0 = time.Now()
	shouldNotChanged(dir, t0)
	shouldNotChanged(filepath.Join(dir, "foo"), t0)
	shouldNotChanged(filepath.Join(dir, "bar"), t0)

	// new dir
	t0 = time.Now()
	name := fmt.Sprintf("%d", rand.Int63())
	err = os.Mkdir(filepath.Join(dir, "foo", name), 0755)
	ce(err)
	w.waitChange(filepath.Join(dir, "foo", name), t0, time.Second*5)
	shouldChanged(dir, t0)
	shouldChanged(filepath.Join(dir, "foo"), t0)
	shouldChanged(filepath.Join(dir, "foo", name), t0)
	shouldNotChanged(filepath.Join(dir, "bar"), t0)

	// new file
	t0 = time.Now()
	f, err = os.CreateTemp(filepath.Join(dir, "foo", name), "*")
	ce(err)
	w.waitChange(f.Name(), t0, time.Second*5)
	shouldChanged(dir, t0)
	shouldChanged(filepath.Join(dir, "foo"), t0)
	shouldChanged(filepath.Join(dir, "foo", name), t0)
	shouldNotChanged(filepath.Join(dir, "bar"), t0)

	// write file
	t0 = time.Now()
	_, err = f.WriteString("foo")
	ce(err)
	ce(f.Sync())
	ce(f.Close())
	w.waitChange(f.Name(), t0, time.Second*5)
	shouldChanged(dir, t0)
	shouldChanged(filepath.Join(dir, "foo"), t0)
	shouldChanged(filepath.Join(dir, "foo", name), t0)
	shouldNotChanged(filepath.Join(dir, "bar"), t0)

	// rename
	t0 = time.Now()
	err = os.Rename(
		f.Name(),
		f.Name()+"foo",
	)
	ce(err)
	_, err = os.Stat(f.Name() + "foo")
	ce(err)
	w.waitChange(f.Name()+"foo", t0, time.Second*5)
	shouldChanged(dir, t0)
	shouldChanged(filepath.Join(dir, "foo"), t0)
	shouldChanged(filepath.Join(dir, "foo", name), t0)
	shouldNotChanged(filepath.Join(dir, "bar"), t0)

	// shuffle
	for i := 0; i < 32; i++ {
		t0 := time.Now()
		op, path1, _, err := shuffle(dir)
		ce(err)
		w.waitChange(path1, t0, time.Second*10)
		ce(err)
		notChanged, err := w.PathNotChanged(path1, t0)
		ce(err)
		if notChanged {
			t.Fatalf("%s %s", op, path1)
		}
	}

	if atomic.LoadInt64(&counter) == 0 {
		t.Fatal()
	}
	if atomic.LoadInt64(&updatedCounter) == 0 {
		t.Fatal()
	}

}

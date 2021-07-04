// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

//go:build linux || windows
// +build linux windows

package fsys

import (
	"context"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/reusee/fsnotify"
)

func sysWatcher(
	ctx context.Context,
	_path string,
	watcher *Watcher,
	tapUpdate []TapUpdatePaths,
	onUpdated []OnUpdatedSpec,
) (
	add func(string) error,
	err error,
) {
	defer he(&err)

	// new fsnotify watcher
	w, err := fsnotify.NewWatcher()
	ce(err)

	go func() {
		defer w.Close()

		delaying := make([]int64, len(onUpdated))

		for {
			select {

			case ev, ok := <-w.Events:
				if !ok {
					panic("watcher error")
				}

				path := filepath.Clean(ev.Name)

				now := time.Now()
				if ev.Op&fsnotify.Create > 0 {
					_, err := watcher.initPath(path, &now)
					if isIgnoreErr(err) {
						break
					}
					ce(err)
				}

				if ev.Op&(fsnotify.Create|
					fsnotify.Remove|
					fsnotify.Rename|
					fsnotify.Write|
					fsnotify.Chmod) > 0 {
					watcher.updatePath(now, path)
				}

				for _, fn := range tapUpdate {
					fn([]string{path})
				}

				for i, spec := range onUpdated {
					i := i
					spec := spec
					if atomic.CompareAndSwapInt64(&delaying[i], 0, 1) {
						time.AfterFunc(spec.MaxDelay, func() {
							spec.Func()
							atomic.StoreInt64(&delaying[i], 0)
						})
					}
				}

			case err, ok := <-w.Errors:
				if !ok {
					panic("waterh error")
				}
				panic(err)

			case <-ctx.Done():
				return

			}
		}

	}()

	add = func(path string) error {
		return w.Add(path)
	}

	return
}

// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package fsys

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/reusee/e5"
)

type WatchOption interface {
	IsWatchOption()
}

type Watch func(
	ctx context.Context,
	path string,
	options ...WatchOption,
) (
	watcher *Watcher,
	err error,
)

type Watcher struct {
	sync.RWMutex
	path         string
	pathParts    []string
	root         *Node
	tapUpdate    TapUpdatePaths
	singleDevice bool
	device       uint64
	add          func(string) error
	Ready        chan struct{}
	traceFiles   bool
}

func (Def) Watch() (
	watch Watch,
) {

	watch = func(
		ctx context.Context,
		path string,
		options ...WatchOption,
	) (
		_ *Watcher,
		err error,
	) {
		defer he(&err)

		path, err = RealPath(path)
		ce(err)
		stat, err := os.Lstat(path)
		ce(err)

		var tapUpdate []TapUpdatePaths
		var singleDevice bool
		var device uint64
		var onInitDone []OnInitDone
		var onUpdated []OnUpdatedSpec
		var trackFiles bool
		for _, option := range options {
			switch option := option.(type) {
			case TapUpdatePaths:
				tapUpdate = append(tapUpdate, option)
			case SingleDevice:
				singleDevice = bool(option)
				if singleDevice {
					device = GetDevice(stat)
				}
			case OnInitDone:
				onInitDone = append(onInitDone, option)
			case OnUpdatedSpec:
				onUpdated = append(onUpdated, option)
			case TrackFiles:
				trackFiles = bool(option)
			default:
				panic(fmt.Errorf("unknown option: %T", option))
			}
		}

		root := new(Node)
		do := make(chan func())

		watcher := &Watcher{
			path:         path,
			pathParts:    strings.Split(path, PathSeparator),
			root:         root,
			singleDevice: singleDevice,
			device:       device,
			Ready:        make(chan struct{}),
			traceFiles:   trackFiles,
		}

		add, err := sysWatcher(
			ctx,
			path,
			watcher,
			tapUpdate,
			onUpdated,
		)
		ce(err)
		watcher.add = add

		// loop
		go func() {

			for {
				select {

				case fn := <-do:
					fn()

				case <-ctx.Done():
					return

				}
			}

		}()

		go func() {
			_, err := watcher.initPath(path, nil)
			ce(err)
			for _, fn := range onInitDone {
				fn()
			}
			close(watcher.Ready)
		}()

		return watcher, nil
	}

	return
}

func (w *Watcher) PathNotChanged(
	path string,
	from time.Time,
) (
	notChanged bool,
	err error,
) {
	defer he(&err)
	modTime, err := w.PathModTime(path)
	ce(err)
	if modTime == nil {
		return false, nil
	}
	return modTime.Before(from), nil
}

func (w *Watcher) PathModTime(path string) (_ *time.Time, err error) {
	defer he(&err)

	path, err = RealPath(path)
	ce(err)

	w.RLock()
	defer w.RUnlock()

	path, err = filepath.Abs(path)
	ce(err)
	parts := strings.Split(path, PathSeparator)

	modTime, ok := w.root.Get(parts)
	if !ok {
		return nil, nil
	}
	return &modTime, nil
}

func (w *Watcher) updatePath(
	t time.Time,
	path string,
) {
	w.Lock()
	defer w.Unlock()

	path = filepath.Clean(path)
	parts := strings.Split(path, PathSeparator)

	w.root.Update(parts, t)
}

func isIgnoreErr(err error) bool {
	if err == nil {
		return false
	}
	if is(err, os.ErrNotExist) || is(err, os.ErrPermission) {
		return true
	}
	var pathErr *os.PathError
	if errors.As(err, &pathErr) {
		errno, ok := pathErr.Err.(syscall.Errno)
		if pathErr.Op == "CreateFile" && ok && errno == 0x7b {
			// invalid file name on Windows
			return true
		}
	}
	return false
}

var initPathCh = make(chan func())

func init() {
	for i := 0; i < runtime.NumCPU(); i++ {
		go func() {
			for fn := range initPathCh {
				fn()
			}
		}()
	}
}

func (w *Watcher) initPath(
	path string,
	modTimeNotBefore *time.Time,
) (
	latest time.Time,
	err error,
) {
	defer he(&err)

	stat, err := os.Stat(path)
	if isIgnoreErr(err) {
		err = nil
		return
	}
	ce(err)

	if stat.Mode()&os.ModeSymlink > 0 {
		return
	}

	if w.singleDevice {
		if GetDevice(stat) != w.device {
			return
		}
	}

	latest = stat.ModTime()
	if modTimeNotBefore != nil && modTimeNotBefore.After(latest) {
		latest = *modTimeNotBefore
	}

	err = w.add(path)
	if isIgnoreErr(err) {
		err = nil
		return
	}
	ce(err,
		e5.Info("add %s", path),
	)

	if stat.IsDir() {
		var f *os.File
		f, err = os.Open(path)
		if isIgnoreErr(err) {
			err = nil
			return
		}
		ce(err,
			e5.Info("open %s", path),
		)
		var latestL sync.Mutex
		for {
			entries, err := f.ReadDir(128)
			if err != nil {
				if is(err, io.EOF) {
					break
				}
				ce(err, e5.Close(f))
			}
			var wg sync.WaitGroup
			errCh := make(chan error, 1)
			for _, entry := range entries {
				if entry.Type()&os.ModeSymlink > 0 {
					continue
				}
				entry := entry
				wg.Add(1)
				select {
				case initPathCh <- func() {
					defer wg.Done()
					l, err := w.initPath(
						filepath.Join(path, entry.Name()),
						modTimeNotBefore,
					)
					if err != nil {
						select {
						case errCh <- err:
						default:
						}
					}
					latestL.Lock()
					if l.After(latest) {
						latest = l
					}
					latestL.Unlock()
				}:
				default:
					l, err := w.initPath(
						filepath.Join(path, entry.Name()),
						modTimeNotBefore,
					)
					ce(err)
					latestL.Lock()
					if l.After(latest) {
						latest = l
					}
					latestL.Unlock()
					wg.Done()
				}
			}
			wg.Wait()
			select {
			case err := <-errCh:
				ce(err)
			default:
			}
		}
		ce(f.Close())
	}

	if w.traceFiles {
		w.updatePath(latest, path)
	} else {
		if stat.IsDir() {
			w.updatePath(latest, path)
		} else {
			w.updatePath(latest, filepath.Dir(path))
		}
	}

	return
}

func (w *Watcher) waitChange(path string, from time.Time, maxDuration time.Duration) {

	path, err := RealPath(path)
	ce(err)
	parts := strings.Split(path, PathSeparator)

	t0 := time.Now()
	for {
		if time.Since(t0) > maxDuration {
			w.Lock()
			w.root.dump("", 0)
			w.Unlock()
			panic(fmt.Errorf("no change: %s from %v", path, from))
		}
		w.Lock()
		modTime, ok := w.root.Get(parts)
		w.Unlock()
		if ok && (modTime.After(from) || modTime.Equal(from)) {
			return
		}
		time.Sleep(time.Millisecond * 50)
	}
}

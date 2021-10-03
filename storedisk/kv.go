// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storedisk

import (
	"io"
	"io/fs"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/reusee/e4"
	"github.com/reusee/june/storekv"
)

var _ storekv.KV = new(Store)

func (s *Store) CostInfo() storekv.CostInfo {
	return storekv.CostInfo{
		Put:    1,
		Delete: 1,
	}
}

func (s *Store) keyToPath(key string) (path string) {
	parts := strings.Split(key, "/")
	hex := parts[len(parts)-1]
	parts = append(
		parts[:len(parts)-1],
		hex[:2],
		hex,
	)
	return filepath.Join(
		append([]string{
			s.dir,
		}, parts...)...,
	)
}

func (s *Store) pathToKey(path string) string {
	parts := strings.Split(path, PathSeparator)
	return strings.Join(parts, "/")
}

func (s *Store) KeyExists(key string) (ok bool, err error) {
	select {
	case <-s.Ctx.Done():
		return false, ErrClosed
	default:
	}
	defer he(&err,
		e4.With(storekv.StringKey(key)),
	)
	path := s.keyToPath(key)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false, nil
	} else {
		ce(err)
	}
	return true, nil
}

func (s *Store) KeyGet(key string, fn func(io.Reader) error) (err error) {
	select {
	case <-s.Ctx.Done():
		return ErrClosed
	default:
	}
	defer he(&err,
		e4.With(storekv.StringKey(key)),
	)
	path := s.keyToPath(key)
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return we.With(e4.With(storekv.StringKey(key)))(ErrKeyNotFound)
	} else {
		ce(err)
	}
	defer f.Close()
	return fn(f)
}

func (s *Store) KeyIter(prefix string, fn func(string) error) (err error) {
	select {
	case <-s.Ctx.Done():
		return ErrClosed
	default:
	}
	defer he(&err,
		e4.NewInfo("prefix: %s", prefix),
	)

	parts := strings.Split(prefix, "/")
	path := filepath.Join(
		append([]string{
			s.dir,
		}, parts...)...,
	)
	paths, err := filepath.Glob(path + "*")
	ce(err)

	for _, path := range paths {
		if err := filepath.WalkDir(path, func(path string, entry fs.DirEntry, err error) (retErr error) {
			defer he(&retErr)
			ce(err)
			if entry.IsDir() {
				return nil
			}
			path, err = filepath.Rel(s.dir, path)
			ce(err)
			parts := strings.Split(path, PathSeparator)
			if strings.Contains(parts[len(parts)-1], ".tmp.") {
				return nil
			}
			if strings.Contains(parts[len(parts)-1], ".deleted") {
				return nil
			}
			path = strings.Join(
				append(parts[:len(parts)-2], parts[len(parts)-1]),
				"/",
			)
			err = fn(path)
			ce(err)
			return nil
		}); is(err, Break) {
			return nil
		} else {
			ce(err)
		}
	}

	return nil
}

func (s *Store) KeyPut(key string, r io.Reader) (err error) {
	select {
	case <-s.Ctx.Done():
		return ErrClosed
	default:
	}
	path := s.keyToPath(key)
	dir := filepath.Dir(path)

	defer he(
		&err,
		e4.NewInfo("path: %s", path),
		e4.NewInfo("dir: %s", dir),
		e4.With(storekv.StringKey(key)),
	)

	if _, ok := s.dirOK.Load(dir); !ok {
		if _, err := os.Stat(dir); err == nil {
			s.dirOK.Store(dir, struct{}{})
		} else if os.IsNotExist(err) {
			err = os.MkdirAll(dir, 0755)
			ce(err)
			s.dirOK.Store(dir, struct{}{})
		} else {
			return err
		}
	}

	tmpPath := path + ".tmp." + strconv.FormatInt(rand.Int63(), 10)
	f, err := os.Create(tmpPath)
	ce(err)
	defer func() {
		if err != nil {
			os.Remove(tmpPath)
		}
	}()
	_, err = io.Copy(f, r)
	ce(err, e4.Close(f))
	if !s.noSync {
		err = f.Sync()
		ce(err)
	}
	ce(f.Close())
	ce(os.Rename(tmpPath, path))

	if atomic.CompareAndSwapInt32(&s.syncPending, 0, 1) {

		done := s.Add()
		timer := time.AfterFunc(time.Second, func() {
			defer done()
			atomic.StoreInt32(&s.syncPending, 0)
			f, err := os.Open(dir)
			if err != nil {
				// skip
				return
			}
			defer f.Close()
			if err := f.Sync(); err != nil {
				// skip
				return
			}
		})

		go func() {
			select {
			case <-s.Ctx.Done():
				// cancel timer
				if !timer.Stop() {
					// func started
				} else {
					// func not start
					done()
				}
			case <-timer.C:
			}
		}()

	}

	return nil
}

func (s *Store) KeyDelete(keys ...string) (err error) {
	select {
	case <-s.Ctx.Done():
		return ErrClosed
	default:
	}
	defer he(&err)
	for _, key := range keys {
		path := s.keyToPath(key)
		if s.softDelete {
			err := os.Rename(path, path+".deleted")
			ce(err,
				e4.With(storekv.StringKey(key)),
				e4.NewInfo("path: %s", path),
			)
		} else {
			err := os.Remove(path)
			ce(err,
				e4.With(storekv.StringKey(key)),
				e4.NewInfo("path: %s", path),
			)
		}
	}
	return nil
}

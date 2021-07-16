// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storepebble

import (
	"bytes"
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"sync/atomic"

	"github.com/cockroachdb/pebble"
	"github.com/cockroachdb/pebble/vfs"
	"github.com/reusee/june/fsys"
	"github.com/reusee/june/naming"
	"github.com/reusee/pr"
	"github.com/reusee/sb"
)

type Store struct {
	*pr.WaitTree
	name    string
	storeID string
	DB      *pebble.DB
}

// create new pebble store
type New func(
	fs vfs.FS,
	dir string,
) (*Store, error)

func (_ Def) New(
	ensureDir fsys.EnsureDir,
	cacheSize CacheSize,
	setRestrictedPath fsys.SetRestrictedPath,
	parentWt *pr.WaitTree,
	machine naming.MachineName,
) New {
	return func(
		fs vfs.FS,
		dir string,
	) (_ *Store, err error) {
		defer he(&err)

		if fs == nil {
			err = ensureDir(dir)
			ce(err)
			err = setRestrictedPath(dir)
			ce(err)
		}
		cache := pebble.NewCache(int64(cacheSize))
		defer cache.Unref()
		db, err := pebble.Open(dir, &pebble.Options{
			FS:                          fs,
			Cache:                       cache,
			Comparer:                    pebbleComparer,
			MaxOpenFiles:                maxOpenFiles,
			MemTableSize:                32 * 1024 * 1024,
			MemTableStopWritesThreshold: 2,
			Logger:                      new(Logger),
			//EventListener: pebble.EventListener{
			//	CompactionBegin: func(info pebble.CompactionInfo) {
			//		pt("compaction: %s\n", info.Reason)
			//		for _, level := range info.Input {
			//			pt("level: %d\n", level.Level)
			//			for _, table := range level.Tables {
			//				pt("size: %d\n", table.Size)
			//			}
			//		}
			//	},
			//	CompactionEnd: func(info pebble.CompactionInfo) {
			//	},
			//},
		})
		ce(err)

		s := &Store{
			name: fmt.Sprintf("pebble%d(%s)",
				atomic.AddInt64(&storeSerial, 1),
				filepath.Base(dir),
			),
			storeID: fmt.Sprintf("pebble(%s, %s)",
				machine,
				dir,
			),
			DB: db,
		}

		s.WaitTree = pr.NewWaitTree(parentWt)
		parentWt.Go(func() {
			<-parentWt.Ctx.Done()
			s.WaitTree.Wait()
			ce(s.DB.Flush())
			ce(s.DB.Close())
		})

		return s, nil
	}

}

var maxOpenFiles = func() int {
	if runtime.GOOS == "darwin" {
		return 256
	}
	return 1024 * 1024
}()

var storeSerial int64

func (s *Store) Name() string {
	return s.name
}

func (s *Store) StoreID() string {
	return s.storeID
}

var pebbleComparer = &pebble.Comparer{
	Compare: func(a, b []byte) int {
		return sb.MustCompareBytes(a, b)
	},
	Equal: func(a, b []byte) bool {
		return bytes.Equal(a, b)
	},
	AbbreviatedKey: func(data []byte) uint64 {
		// encoded token kind
		return uint64(data[0])
	},
	Separator: func(dst, a, b []byte) []byte { // NOCOVER
		return a
	},
	Successor: func(dst, a []byte) []byte { // NOCOVER
		return a
	},
}

func catchErr(errp *error, errs ...error) {
	p := recover()
	if p == nil {
		return
	}
	if e, ok := p.(error); ok {
		for _, err := range errs {
			if errors.Is(e, err) {
				if errp != nil {
					*errp = e
				}
				return
			}
		}
	}
	panic(p)
}

func (s *Store) Sync() (err error) {
	return nil
}

var writeOptions = &pebble.WriteOptions{
	Sync: false,
}

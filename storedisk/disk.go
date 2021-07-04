// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storedisk

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/reusee/e4"
	"github.com/reusee/ling/v2/fsys"
	"github.com/reusee/ling/v2/naming"
	"github.com/reusee/ling/v2/sys"
	"github.com/reusee/pr"
)

type Store struct {
	*pr.WaitTree
	name        string
	storeID     string
	id          string
	ensureDir   fsys.EnsureDir
	dirOK       sync.Map
	dir         string
	syncPending int32
	softDelete  SoftDelete
	noSync      bool
	closed      chan struct{}
	closeOnce   sync.Once
}

type NewOption interface {
	IsNewOption()
}

type NewOptions []NewOption

func (e NewOptions) Error() string {
	var b strings.Builder
	for i, option := range e {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(fmt.Sprintf("option %#v", option))
	}
	return b.String()
}

type New func(string, ...NewOption) (*Store, error)

func (_ Def) New(
	ensureDir fsys.EnsureDir,
	setRestrictedPath fsys.SetRestrictedPath,
	isTesting sys.Testing,
	rootWaitTree *pr.WaitTree,
	machine naming.MachineName,
) New {

	noSync := bool(isTesting && runtime.GOOS == "darwin")

	return func(dir string, options ...NewOption) (_ *Store, err error) {
		defer he(&err,
			e4.NewInfo("dir: %s", dir),
			e4.With(NewOptions(options)),
		)

		err = ensureDir(dir)
		ce(err)
		err = setRestrictedPath(dir)
		ce(err)
		wt := pr.NewWaitTree(rootWaitTree, nil)
		store := &Store{
			name: fmt.Sprintf("disk%d(%s)",
				atomic.AddInt64(&serial, 1),
				filepath.Base(dir),
			),
			storeID: fmt.Sprintf("disk(%s, %s)",
				machine,
				dir,
			),
			WaitTree:  wt,
			dir:       dir,
			ensureDir: ensureDir,
			noSync:    noSync,
			closed:    make(chan struct{}),
		}
		for _, option := range options {
			switch option := option.(type) {
			case SoftDelete:
				store.softDelete = option
			default:
				panic(fmt.Errorf("unknown option: %T", option))
			}
		}
		return store, nil
	}
}

var serial int64

func (s *Store) Name() string {
	return s.name
}

func (s *Store) StoreID() string {
	return s.storeID
}

func (s *Store) Close() error {
	s.closeOnce.Do(func() {
		s.Cancel()
		close(s.closed)
		s.Wait()
		s.Done()
	})
	return nil
}

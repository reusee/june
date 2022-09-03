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

	"github.com/reusee/e5"
	"github.com/reusee/june/fsys"
	"github.com/reusee/june/naming"
	"github.com/reusee/june/sys"
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

type New func(
	wt *pr.WaitTree,
	path string,
	options ...NewOption,
) (*Store, error)

func (_ Def) New(
	ensureDir fsys.EnsureDir,
	setRestrictedPath fsys.SetRestrictedPath,
	isTesting sys.Testing,
	machine naming.MachineName,
) New {

	noSync := bool(isTesting && runtime.GOOS == "darwin")

	return func(
		parentWt *pr.WaitTree,
		dir string,
		options ...NewOption,
	) (_ *Store, err error) {
		defer he(&err,
			e5.NewInfo("dir: %s", dir),
			e5.With(NewOptions(options)),
		)

		err = ensureDir(dir)
		ce(err)
		err = setRestrictedPath(dir)
		ce(err)
		store := &Store{
			WaitTree: parentWt,
			name: fmt.Sprintf("disk%d(%s)",
				atomic.AddInt64(&serial, 1),
				filepath.Base(dir),
			),
			storeID: fmt.Sprintf("disk(%s, %s)",
				machine,
				dir,
			),
			dir:       dir,
			ensureDir: ensureDir,
			noSync:    noSync,
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

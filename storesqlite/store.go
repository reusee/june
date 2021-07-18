// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storesqlite

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/reusee/june/naming"
	"github.com/reusee/pr"
)

type Store struct {
	*pr.WaitTree
	sync.RWMutex
	name    string
	storeID string
	DB      *sql.DB
}

type New func(
	path string,
) (*Store, error)

func (_ Def) New(
	parentWt *pr.WaitTree,
	machine naming.MachineName,
) New {
	return func(
		path string,
	) (_ *Store, err error) {
		defer he(&err)

		db, err := sql.Open("sqlite3", "file:"+path)
		ce(err)

		_, err = db.Exec(`
      create table if not exists kv (
        kind integer(1),
        key text,
        value blob
      )
    `)
		ce(err)

		s := &Store{
			name: fmt.Sprintf("sqlite%d(%s)",
				time.Now().UnixNano(),
				filepath.Base(path),
			),
			storeID: fmt.Sprintf("pebble(%s, %s)",
				machine,
				path,
			),
			DB: db,
		}

		s.WaitTree = pr.NewWaitTree(parentWt)
		parentWt.Go(func() {
			<-parentWt.Ctx.Done()
			s.WaitTree.Wait()
			ce(db.Close())
		})

		return s, nil
	}
}

func (s *Store) Name() string {
	return s.name
}

func (s *Store) StoreID() string {
	return s.storeID
}

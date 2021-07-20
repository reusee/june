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
	"github.com/reusee/e4"
	"github.com/reusee/june/fsys"
	"github.com/reusee/june/naming"
	"github.com/reusee/june/storemem"
	"github.com/reusee/pr"
)

type Store struct {
	*pr.WaitTree
	name    string
	storeID string

	cond     *sync.Cond
	numRead  int
	numWrite int
	DB       *sql.DB
	mem      sync.Map
	dirty    chan struct{}
}

type New func(
	path string,
) (*Store, error)

func (_ Def) New(
	parentWt *pr.WaitTree,
	machine naming.MachineName,
	newMem storemem.New,
	setRestrictedPath fsys.SetRestrictedPath,
) New {
	return func(
		path string,
	) (_ *Store, err error) {
		defer he(&err)

		db, err := sql.Open("sqlite3", "file:"+path)
		ce(err)

		ce(setRestrictedPath(path))

		_, err = db.Exec(`
      create table if not exists kv (
        kind integer(1),
        key text unique,
        value blob
      )
    `)
		ce(err)

		s := &Store{
			cond: sync.NewCond(new(sync.Mutex)),
			name: fmt.Sprintf("sqlite%d(%s)",
				time.Now().UnixNano(),
				filepath.Base(path),
			),
			storeID: fmt.Sprintf("sqlite(%s, %s)",
				machine,
				path,
			),
			DB:    db,
			dirty: make(chan struct{}, 1),
		}

		s.WaitTree = pr.NewWaitTree(parentWt)
		parentWt.Go(func() {
			<-parentWt.Ctx.Done()
			s.WaitTree.Wait()
			ce(db.Close())
		})

		s.WaitTree.Go(s.sync)

		return s, nil
	}
}

func (s *Store) Name() string {
	return s.name
}

func (s *Store) StoreID() string {
	return s.storeID
}

func (s *Store) lockRead() func() {
	s.cond.L.Lock()
	for s.numWrite > 0 {
		s.cond.Wait()
	}
	s.numRead++
	s.cond.L.Unlock()
	return func() {
		s.cond.L.Lock()
		s.numRead--
		s.cond.L.Unlock()
		s.cond.Broadcast()
	}
}

func (s *Store) sync() {

	sync := func() (err error) {
		defer he(&err)

		s.cond.L.Lock()
		for s.numRead > 0 || s.numWrite > 0 {
			s.cond.Wait()
		}
		s.numWrite++
		s.cond.L.Unlock()
		defer func() {
			s.cond.L.Lock()
			s.numWrite--
			s.cond.L.Unlock()
			s.cond.Broadcast()
		}()

		tx, err := s.DB.Begin()
		ce(err)
		defer he(&err, e4.Do(func() {
			tx.Rollback()
		}))

		// put
		s.mem.Range(func(k, v any) bool {
			key := k.(string)

			if v == nil {
				// delete
				_, err = tx.Exec(`
          delete from kv
          where kind = ?
          and key = ?
          `,
					Kv,
					key,
				)
				ce(err)

			} else {
				// put
				ce(s.put(tx, key, v.([]byte)))
			}

			s.mem.Delete(key)

			return true
		})

		ce(tx.Commit())

		return nil
	}

	for {
		select {
		case <-s.dirty:
			ce(sync())
		case <-s.Ctx.Done():
			ce(sync())
			return
		}
	}
}

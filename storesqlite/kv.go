// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storesqlite

import (
	"bytes"
	"database/sql"
	"errors"
	"io"
	"strings"

	"github.com/reusee/e4"
	"github.com/reusee/june/storekv"
)

const (
	Kv = iota
	Idx
)

var _ storekv.KV = new(Store)

func (s *Store) KeyExists(key string) (ok bool, err error) {
	defer he(&err)
	done := s.Add()
	defer done()

	defer s.lockRead()()

	v, ok := s.mem.Load(key)
	if ok {
		if v == nil {
			return false, nil
		} else {
			return true, nil
		}
	}

	var exists bool
	ce(s.DB.QueryRow(`
    select exists (
      select 1 from kv
      where kind = ?
      and key = ?
    )
    `,
		Kv,
		key,
	).Scan(&exists))

	return exists, nil
}

func (s *Store) KeyGet(key string, fn func(io.Reader) error) (err error) {
	defer he(&err)
	done := s.Add()
	defer done()

	defer s.lockRead()()

	v, ok := s.mem.Load(key)
	if ok {
		if v == nil {
			return we(storekv.ErrKeyNotFound,
				e4.With(storekv.StringKey(key)),
			)
		} else {
			return fn(bytes.NewReader(v.([]byte)))
		}
	}

	var data []byte
	err = s.DB.QueryRow(`
    select value from kv
    where kind = ?
    and key = ?
    `,
		Kv,
		key,
	).Scan(&data)
	if errors.Is(err, sql.ErrNoRows) {
		return we(storekv.ErrKeyNotFound,
			e4.With(storekv.StringKey(key)),
		)
	}
	ce(err)

	if fn != nil {
		err = fn(bytes.NewReader(data))
		ce(err)
	}

	return nil
}

func (s *Store) KeyPut(key string, r io.Reader) (err error) {
	defer he(&err)
	done := s.Add()
	defer done()

	// try lock write
	ok := false
	s.cond.L.Lock()
	if s.numRead == 0 && s.numWrite == 0 {
		s.numWrite++
		ok = true
	}
	s.cond.L.Unlock()
	if ok {
		defer func() {
			s.cond.L.Lock()
			s.numWrite--
			s.cond.L.Unlock()
			s.cond.Broadcast()
		}()

		var bs []byte
		if b, ok := r.(interface {
			Bytes() []byte
		}); ok {
			bs = b.Bytes()
		} else {
			bs, err = io.ReadAll(r)
			ce(err)
		}

		var tx *sql.Tx
		tx, err = s.DB.Begin()
		ce(err)
		defer he(&err, e4.Do(func() {
			tx.Rollback()
		}))
		ce(s.put(tx, key, bs))
		ce(tx.Commit())

		return
	}

	defer s.lockRead()()

	bs, err := io.ReadAll(r)
	ce(err)

	s.mem.Store(key, bs)

	select {
	case s.dirty <- struct{}{}:
	default:
	}

	return nil
}

func (s *Store) KeyDelete(keys ...string) (err error) {
	defer he(&err)
	done := s.Add()
	defer done()

	defer s.lockRead()()

	for _, key := range keys {
		s.mem.Store(key, nil)
	}

	select {
	case s.dirty <- struct{}{}:
	default:
	}

	return
}

func (s *Store) KeyIter(prefix string, fn func(key string) error) (err error) {
	defer he(&err)
	done := s.Add()
	defer done()

	defer s.lockRead()()

	visited := make(map[string]struct{})
	isBreak := false
	s.mem.Range(func(k, v any) bool {
		key := k.(string)
		visited[key] = struct{}{}
		if v == nil {
			return true
		}
		if !strings.HasPrefix(key, prefix) {
			return true
		}
		err := fn(key)
		if errors.Is(err, storekv.Break) {
			isBreak = true
			return false
		}
		ce(err)
		return true
	})
	if isBreak {
		return nil
	}

	rows, err := s.DB.Query(`
    select key
    from kv
    where kind = ?
    and key like ?
    `,
		Kv,
		prefix+"%",
	)
	ce(err)
	for rows.Next() {
		var key string
		ce(rows.Scan(&key))
		if _, ok := visited[key]; ok {
			continue
		}
		err = fn(key)
		if errors.Is(err, storekv.Break) {
			break
		}
		ce(err)
	}
	ce(rows.Err())
	rows.Close()

	return
}

func (s *Store) put(tx *sql.Tx, key string, value []byte) error {
	var exists bool
	if err := tx.QueryRow(`
          select exists (
            select 1 from kv
            where kind = ?
            and key = ?
          )
          `,
		Kv,
		key,
	).Scan(&exists); err != nil {
		return err
	}

	if !exists {
		_, err := tx.Exec(`
            insert into kv
            (kind, key, value)
            values 
            (?, ?, ?)
            `,
			Kv,
			key,
			value,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

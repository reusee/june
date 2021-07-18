// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storesqlite

import (
	"bytes"
	"database/sql"
	"errors"
	"io"

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

	s.RLock()
	defer s.RUnlock()

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

	s.RLock()
	defer s.RUnlock()

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

	s.Lock()
	defer s.Unlock()

	tx, err := s.DB.Begin()
	ce(err)
	defer he(&err, e4.Do(func() {
		tx.Rollback()
	}))

	var exists bool
	ce(tx.QueryRow(`
    select exists (
      select 1 from kv
      where kind = ?
      and key = ?
    )
    `,
		Kv,
		key,
	).Scan(&exists))
	if exists {
		return tx.Rollback()
	}

	var bs []byte
	if b, ok := r.(interface {
		Bytes() []byte
	}); ok {
		bs = b.Bytes()
	} else {
		bs, err = io.ReadAll(r)
		ce(err)
	}

	_, err = tx.Exec(`
    insert into kv
    (kind, key, value)
    values 
    (?, ?, ?)
    `,
		Kv,
		key,
		bs,
	)
	ce(err)
	ce(tx.Commit())

	return
}

func (s *Store) KeyDelete(keys ...string) (err error) {
	defer he(&err)
	done := s.Add()
	defer done()

	s.Lock()
	defer s.Unlock()

	tx, err := s.DB.Begin()
	ce(err)
	defer he(&err, e4.Do(func() {
		tx.Rollback()
	}))

	for _, key := range keys {
		_, err = tx.Exec(`
      delete from kv
      where kind = ?
      and key = ?
      `,
			Kv,
			key,
		)
		ce(err)
	}

	ce(tx.Commit())

	return
}

func (s *Store) KeyIter(prefix string, fn func(key string) error) (err error) {
	defer he(&err)
	done := s.Add()
	defer done()

	s.RLock()
	defer s.RUnlock()

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

// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package vars

import (
	"bytes"
	"errors"
	"fmt"
	"runtime"

	"github.com/cockroachdb/pebble"
	"github.com/reusee/e4"
	"github.com/reusee/june/fsys"
	"github.com/reusee/june/storepebble"
	"github.com/reusee/pr"
	"github.com/reusee/sb"
)

type VarsDir string

type VarsStore struct {
	db *pebble.DB
}

func (_ Def) VarsStore(
	ensureDir fsys.EnsureDir,
	dir VarsDir,
	setRestrictedPath fsys.SetRestrictedPath,
	wt *pr.WaitTree,
) *VarsStore {

	ce(ensureDir(string(dir)))
	ce(setRestrictedPath(string(dir)))

	db, err := pebble.Open(string(dir), &pebble.Options{
		MaxOpenFiles:                maxOpenFiles,
		MemTableSize:                8 * 1024 * 1024,
		MemTableStopWritesThreshold: 2,
		Logger:                      new(storepebble.Logger),
	})
	ce(err)

	done := wt.Add()
	go func() {
		defer done()
		<-wt.Ctx.Done()
		var err error
		defer catchErr(&err, pebble.ErrClosed)
		ce(db.Close())
	}()

	return &VarsStore{
		db: db,
	}
}

var maxOpenFiles = func() int {
	if runtime.GOOS == "darwin" {
		return 256
	}
	return 1024 * 1024
}()

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

type Get func(key string, target any) error

func (_ Def) Get(
	store *VarsStore,
) Get {

	return func(key string, target any) (err error) {
		defer catchErr(&err, pebble.ErrClosed)
		bs, c, err := store.db.Get([]byte(key))
		if err != nil {
			if is(err, pebble.ErrNotFound) {
				return we(ErrNotFound, e4.With(&NotFound{
					Key: key,
				}))
			}
			return err
		}
		defer c.Close()
		if err := sb.Copy(
			sb.Decode(bytes.NewReader(bs)),
			sb.Unmarshal(target),
		); err != nil {
			return err
		}
		return nil
	}

}

type NotFound struct {
	Key string
}

func (n *NotFound) Error() string {
	return fmt.Sprintf("var not found: %s", n.Key)
}

var ErrNotFound = errors.New("not found")

type Set func(key string, value any) error

func (_ Def) Set(
	store *VarsStore,
) Set {

	return func(key string, value any) (err error) {
		defer catchErr(&err, pebble.ErrClosed)
		buf := new(bytes.Buffer)
		if err := sb.Copy(
			sb.Marshal(value),
			sb.Encode(buf),
		); err != nil {
			return err
		}
		if err := store.db.Set([]byte(key), buf.Bytes(), &pebble.WriteOptions{
			Sync: false,
		}); err != nil {
			return err
		}
		return nil
	}

}

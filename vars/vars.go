// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package vars

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"runtime"

	"github.com/cockroachdb/pebble"
	"github.com/reusee/e5"
	"github.com/reusee/june/fsys"
	"github.com/reusee/june/storepebble"
	"github.com/reusee/pr"
	"github.com/reusee/sb"
)

type VarsSpec func() (
	dir string,
	ctx context.Context,
)

type VarsStore struct {
	ctx context.Context
	db  *pebble.DB
}

func (Def) VarsStore(
	ensureDir fsys.EnsureDir,
	spec VarsSpec,
	setRestrictedPath fsys.SetRestrictedPath,
) *VarsStore {

	dir, ctx := spec()

	ce(ensureDir(string(dir)))
	ce(setRestrictedPath(string(dir)))

	db, err := pebble.Open(string(dir), &pebble.Options{
		MaxOpenFiles:                maxOpenFiles,
		MemTableSize:                8 * 1024 * 1024,
		MemTableStopWritesThreshold: 2,
		Logger:                      new(storepebble.Logger),
	})
	ce(err)

	ctx, wg := pr.WithWaitGroup(ctx)

	wg.Parent().Go(func() {
		<-ctx.Done()
		wg.Wait()
		var err error
		defer catchErr(&err, pebble.ErrClosed)
		ce(db.Close())
	})

	return &VarsStore{
		ctx: ctx,
		db:  db,
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

type Get func(ctx context.Context, key string, target any) error

func (Def) Get(
	store *VarsStore,
) Get {

	return func(ctx context.Context, key string, target any) (err error) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		defer catchErr(&err, pebble.ErrClosed)
		bs, c, err := store.db.Get([]byte(key))
		if err != nil {
			if is(err, pebble.ErrNotFound) {
				return we.With(
					e5.With(&NotFound{Key: key}),
				)(ErrNotFound)
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

type Set func(ctx context.Context, key string, value any) error

func (Def) Set(
	store *VarsStore,
) Set {

	return func(ctx context.Context, key string, value any) (err error) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

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

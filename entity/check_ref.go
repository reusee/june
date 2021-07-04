// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package entity

import (
	"context"
	"fmt"

	"github.com/reusee/e4"
	"github.com/reusee/june/opts"
	"github.com/reusee/june/sys"
	"github.com/reusee/pr"
	"github.com/reusee/sb"
)

type CheckRefOption interface {
	IsCheckRefOption()
}

type CheckRef func(
	options ...CheckRefOption,
) error

func (_ Def) CheckRef(
	store Store,
	rootCtx context.Context,
	parallel sys.Parallel,
) CheckRef {

	return func(options ...CheckRefOption) (err error) {
		defer he(&err)

		var tapKey opts.TapKey
		for _, option := range options {
			switch option := option.(type) {
			case opts.TapKey:
				tapKey = option
			default:
				panic(fmt.Errorf("bad option: %T", option))
			}
		}

		ctx, cancel := context.WithCancel(rootCtx)
		defer cancel()

		put, wait := pr.Consume(ctx, int(parallel), func(i int, v any) (err error) {
			defer he(&err)
			key := v.(Key)
			var summary Summary
			ce(store.Read(key, func(s sb.Stream) error {
				return sb.Copy(s, sb.Unmarshal(&summary))
			}))
			ce(summary.checkRef(store))
			if tapKey != nil {
				tapKey(summary.Key)
			}
			return nil
		})

		ce(store.IterKeys(NSSummary, func(key Key) error {
			put(key)
			return nil
		}))
		ce(wait(true))

		return nil
	}
}

func (s *Summary) checkRef(store Store) (err error) {
	defer he(&err)

	// Key
	ok, err := store.Exists(s.Key)
	ce(err)
	if !ok {
		var typeName string
		if s.Indexes != nil {
			for _, idx := range *s.Indexes {
				if _, ok := idx.Type.(idxType); ok {
					typeName = idx.Tuple[0].(string)
				}
			}
		}
		return we(
			ErrKeyNotFound,
			e4.NewInfo("lost key: %s", s.Key),
			e4.NewInfo("entity type: %s", typeName),
		)
	}

	checkKey := func(idx IndexEntry, key Key) (err error) {
		defer he(&err)
		ok, err := store.Exists(key)
		ce(err)
		if !ok {
			return we(
				ErrKeyNotFound,
				e4.NewInfo("entity key: %s", s.Key),
				e4.NewInfo("index tuple: %+v", idx),
				e4.NewInfo("lost key: %s", key),
			)
		}
		return nil
	}

	// Indexes
	if s.Indexes != nil {
		for _, idx := range *s.Indexes {
			for _, elem := range idx.Tuple {
				// index type objects (the first elements of index.Tuple) may contain empty key
				// skip them
				if key, ok := elem.(Key); ok && key.Valid() {
					ce(checkKey(idx, key))
				}
			}
			if idx.Key != nil {
				ce(checkKey(idx, *idx.Key))
			}
			if idx.Path != nil && len(*idx.Path) > 0 {
				ce(checkKey(idx, (*idx.Path)[len(*idx.Path)-1]))
			}
		}
	}

	// Subs
	for _, sub := range s.Subs {
		ce(sub.checkRef(store))
	}

	// ReferedKeys
	for _, key := range s.ReferedKeys {
		ok, err := store.Exists(key)
		ce(err)
		if !ok {
			return we(ErrKeyNotFound, e4.With(key))
		}
	}

	return nil
}

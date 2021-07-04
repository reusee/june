// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package entity

import (
	"context"
	"fmt"

	"github.com/reusee/june/sys"
	"github.com/reusee/pp"
	"github.com/reusee/pr"
	"github.com/reusee/sb"
)

type DeleteIndex func(
	predict func(sb.Stream) (*IndexEntry, error),
	options ...DeleteIndexOption,
) error

type DeleteIndexOption interface {
	IsDeleteIndexOption()
}

func (_ Def) DeleteIndex(
	index Index,
	rootCtx context.Context,
	parallel sys.Parallel,
) DeleteIndex {

	return func(
		predict func(sb.Stream) (*IndexEntry, error),
		options ...DeleteIndexOption,
	) (err error) {
		defer he(&err)

		var tapDelete TapDeleteIndex
		for _, option := range options {
			switch option := option.(type) {
			case TapDeleteIndex:
				tapDelete = option
			default:
				panic(fmt.Errorf("bad option: %T", option))
			}
		}

		iter, closer, err := index.Iter(
			nil,
			nil,
			Asc,
		)
		ce(err)
		defer closer.Close()

		ctx, cancel := context.WithCancel(rootCtx)
		defer cancel()
		put, wait := pr.Consume(ctx, int(parallel), func(_ int, v any) (err error) {
			defer he(&err)
			entry := v.(IndexEntry)
			ce(index.Delete(entry))
			return nil
		})

		ce(pp.Copy(iter, pp.Tap(func(v any) (err error) {
			stream := v.(sb.Stream)
			defer he(&err)

			entry, err := predict(stream)
			ce(err)

			if entry != nil {
				if tapDelete != nil {
					tapDelete(*entry)
				}
				put(*entry)
			}

			return nil
		})))
		ce(closer.Close())

		ce(wait(true))

		return
	}
}

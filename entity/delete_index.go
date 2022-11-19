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
	ctx context.Context,
	predict func(sb.Stream) (*IndexEntry, error),
	options ...DeleteIndexOption,
) error

type DeleteIndexOption interface {
	IsDeleteIndexOption()
}

func (_ Def) DeleteIndex(
	index Index,
	parallel sys.Parallel,
) DeleteIndex {

	return func(
		ctx context.Context,
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
			ctx,
			nil,
			nil,
			Asc,
		)
		ce(err)
		defer closer.Close()

		ctx, wg := pr.WithWaitGroup(ctx)
		defer wg.Cancel()
		put, wait := pr.Consume(ctx, int(parallel), func(_ int, v any) (err error) {
			defer he(&err)
			entry := v.(IndexEntry)
			ce(index.Delete(ctx, entry))
			return nil
		})

		ce(pp.Copy(iter, pp.Tap(func(v any) (err error) {
			stream := v.(sb.Stream)
			defer he(&err)

			tupleToken, err := stream.Next()
			ce(err)
			if tupleToken.Kind != sb.KindTuple {
				panic("bad index stream")
			}
			firstToken, err := stream.Next()
			ce(err)
			if firstToken.Kind != sb.KindString {
				// not Entry
				return nil
			}

			s := sb.ConcatStreams(
				sb.Tokens{*tupleToken, *firstToken}.Iter(),
				stream,
			)

			entry, err := predict(s)
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

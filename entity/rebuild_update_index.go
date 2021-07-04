// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package entity

import (
	"context"
	"fmt"
	"reflect"
	"sync/atomic"

	"github.com/reusee/june/sys"
	"github.com/reusee/pr"
	"github.com/reusee/sb"
)

type RebuildIndex func(
	options ...IndexOption,
) (
	n int64,
	err error,
)

type UpdateIndex func(
	options ...IndexOption,
) (
	n int64,
	err error,
)

type IndexOption interface {
	IsIndexOption()
}

type WithIndexSaveOptions []IndexSaveOption

func (_ WithIndexSaveOptions) IsIndexOption() {}

func (_ Def) IndexFuncs(
	store Store,
	saveSummary SaveSummary,
	index Index,
	rootCtx context.Context,
	parallel sys.Parallel,
) (rebuild RebuildIndex, update UpdateIndex) {

	resave := func(
		ignore func(summaryKey Key) (bool, error),
		options ...IndexOption,
	) (n int64, err error) {
		defer he(&err)

		var tapKey TapKey
		var saveOptions []IndexSaveOption
		for _, option := range options {
			switch option := option.(type) {
			case TapKey:
				tapKey = option
			case WithIndexSaveOptions:
				saveOptions = append(saveOptions, option...)
			default:
				panic(fmt.Errorf("unknown option %T", option))
			}
		}

		ctx, cancel := context.WithCancel(rootCtx)
		defer cancel()

		put, wait := pr.Consume(ctx, int(parallel), func(i int, v any) (err error) {
			defer he(&err)

			key := v.(Key)

			if tapKey != nil {
				tapKey(key)
			}

			// check ignore
			if ignore != nil {
				shouldIgnore, err := ignore(key)
				ce(err)
				if shouldIgnore {
					return nil
				}
			}

			// fetch
			var summary Summary
			ce(
				store.Read(key, func(s sb.Stream) error {
					err := sb.Copy(
						s,
						sb.UnmarshalValue(sb.Ctx{
							Unmarshal:                   sb.UnmarshalValue,
							DisallowUnknownStructFields: true,
						}, reflect.ValueOf(&summary), nil),
					)
					return err
				}),
			)
			if !summary.Valid() {
				return fmt.Errorf("summary not valid")
			}

			// save
			ce(saveSummary(&summary, false, WithIndexSaveOptions(saveOptions)))
			atomic.AddInt64(&n, 1)

			return nil
		})

		ce(store.IterKeys(NSSummary, func(key Key) error {
			put(key)
			return nil
		}))
		ce(wait(true))

		return n, nil
	}

	rebuild = func(
		options ...IndexOption,
	) (n int64, err error) {
		return resave(nil, options...)
	}

	update = func(
		options ...IndexOption,
	) (n int64, err error) {
		return resave(func(key Key) (_ bool, err error) {
			defer he(&err)
			var c int
			ce(Select(
				index,
				MatchEntry(IdxSummaryOf, key),
				Count(&c),
				Limit(1),
			))
			if c > 0 {
				return true, nil
			}
			return
		}, options...)
	}

	return
}

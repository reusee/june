// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package entity

import (
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"

	"github.com/reusee/june/index"
	"github.com/reusee/june/naming"
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
	sel index.SelectIndex,
	wt *pr.WaitTree,
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

		wt := pr.NewWaitTree(wt)
		defer wt.Cancel()
		put, wait := pr.Consume(wt, int(parallel), func(i int, v any) (err error) {
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
		return resave(func(summaryKey Key) (_ bool, err error) {
			defer he(&err)

			// check existence
			var c int
			var key Key
			ce(sel(
				MatchEntry(IdxSummaryOf, summaryKey),
				Count(&c),
				Limit(1),
				TapEntry(func(e IndexEntry) {
					key = *e.Key
				}),
			))
			if c == 0 {
				return false, nil
			}

			// check version
			var typeName string
			ce(sel(
				MatchPreEntry(key, IdxType),
				TapPre(func(t string) {
					typeName = t
				}),
			))
			ver, err := getVersion(typeName)
			ce(err)
			if ver != nil {
				// get saved version
				var savedVersion int64
				var c int
				ce(sel(
					MatchPreEntry(key, IdxVersion),
					TapPre(func(v int64) {
						savedVersion = v
					}),
					Count(&c),
				))
				if c == 0 {
					// no version info
					return false, nil
				}
				if savedVersion != *ver {
					return false, nil
				}
			}

			return true, nil
		}, options...)
	}

	return
}

var nameToVersion sync.Map

func getVersion(
	name string,
) (ret *int64, err error) {

	if v, ok := nameToVersion.Load(name); ok {
		return v.(*int64), nil
	}

	defer func() {
		if err == nil {
			nameToVersion.Store(name, ret)
		}
	}()

	t := naming.GetType(name)
	if t == nil {
		return nil, nil
	}
	if !(*t).Implements(hasIndexType) {
		return nil, nil
	}

	_, ver, err := reflect.New(*t).Elem().Interface().(HasIndex).EntityIndexes()
	if err != nil {
		return nil, err
	}

	return &ver, nil
}

// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package entity

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/reusee/june/index"
	"github.com/reusee/june/opts"
	"github.com/reusee/june/sys"
	"github.com/reusee/pr"
	"github.com/reusee/sb"
)

type Push func(
	to Store,
	toIndex IndexManager,
	keys []Key,
	options ...PushOption,
) error

type PushOption interface {
	IsPushOption()
}

type TapPushCheckSummary func(summaryKey Key)

func (_ TapPushCheckSummary) IsPushOption() {}

type TapPushSave func(summaryKey Key, summary *Summary)

func (_ TapPushSave) IsPushOption() {}

func (_ Def) Push(
	scope Scope,
	selIndex index.SelectIndex,
	store Store,
	rootCtx context.Context,
	fetch Fetch,
	parallel sys.Parallel,
) Push {

	return func(
		to Store,
		toIndex IndexManager,
		keys []Key,
		options ...PushOption,
	) (err error) {
		defer he(&err)

		p := int(parallel)

		var tapCheck []TapPushCheckSummary
		var tapSave []TapPushSave
		for _, option := range options {
			switch option := option.(type) {
			case TapPushCheckSummary:
				tapCheck = append(tapCheck, option)
			case TapPushSave:
				tapSave = append(tapSave, option)
			case Parallel:
				p = int(option)
			default:
				panic(fmt.Errorf("bad option: %T", option))
			}
		}

		// to Scope
		toDecls := []any{
			func() Store {
				return to
			},
		}
		if toIndex != nil {
			toDecls = append(toDecls, func() IndexManager {
				return toIndex
			})
		}
		toScope := scope.Sub(toDecls...)

		// func to check summary key existence
		var keyExisted func(summaryKey Key) (bool, error)

		if toIndex == nil {
			// iterate keys
			keySet := make(map[Key]struct{})
			ce(to.IterKeys(NSSummary, func(key Key) error {
				keySet[key] = struct{}{}
				return nil
			}))
			keyExisted = func(key Key) (bool, error) {
				_, ok := keySet[key]
				return ok, nil
			}

		} else {
			// by index
			var toSelectIndex index.SelectIndex
			toScope.Assign(&toSelectIndex)
			keyExisted = func(key Key) (bool, error) {
				var n int
				if err := toSelectIndex(
					MatchEntry(IdxPairSummaryObject, key),
					Count(&n),
				); err != nil {
					return false, err
				}
				return n > 0, nil
			}
		}

		// func to iter root keys
		var iterKeys func(
			func(summaryKey Key) error,
		)

		if len(keys) > 0 {
			// some keys
			iterKeys = func(
				fn func(summaryKey Key) error,
			) {
				for _, key := range keys {
					if key.Namespace != NSSummary {
						ce(selIndex(
							MatchEntry(IdxPairObjectSummary, key),
							Tap(func(_ Key, summaryKey Key) {
								ce(fn(summaryKey))
							}),
						))
					}
				}
			}

		} else {
			// all keys
			iterKeys = func(
				fn func(summaryKey Key) error,
			) {
				ce(store.IterKeys(NSSummary, func(key Key) error {
					return fn(key)
				}))
			}
		}

		var toSaveSummary SaveSummary
		toScope.Assign(&toSaveSummary)
		// save summary and object
		save := func(summaryKey Key, summary *Summary) (err error) {
			defer he(&err)

			for _, fn := range tapSave {
				fn(summaryKey, summary)
			}

			// save object
			ce(store.Read(summary.Key, func(stream sb.Stream) (err error) {
				defer he(&err)
				get, put := bytesPool.Getter()
				defer put()
				res, err := to.Write(summary.Key.Namespace, stream, opts.NewBytesBuffer(func() *bytes.Buffer {
					buf := get().(*bytes.Buffer)
					if buf.Len() > 0 {
						buf.Reset()
					}
					return buf
				}))
				ce(err)
				if res.Key != summary.Key {
					return we(fmt.Errorf("bad write: %s", summary.Key))
				}
				return
			}))

			// save summary
			var retKey Key
			ce(toSaveSummary(
				summary,
				false,
				TapKey(func(k Key) {
					retKey = k
				}),
			))
			if retKey != summaryKey {
				return we(fmt.Errorf("bad summary write: %s", summaryKey))
			}

			return
		}

		type Proc func() error

		// workers
		ctx, cancel := context.WithCancel(rootCtx)
		defer cancel()
		put, wait := pr.Consume(ctx, p, func(_ int, v any) error {
			proc := v.(Proc)
			if proc == nil {
				return nil
			}
			return proc()
		})

		var pushing sync.Map

		// check summary key
		var check func(summary Key, cont Proc) Proc
		check = func(summaryKey Key, cont Proc) Proc {
			return func() (err error) {
				defer he(&err)

				// check to store existence
				existed, err := keyExisted(summaryKey)
				ce(err)
				if existed {
					put(cont)
					return nil
				}

				// check if pushing
				if _, ok := pushing.Load(summaryKey); ok {
					put(cont)
					return nil
				}
				pushing.Store(summaryKey, struct{}{})

				for _, fn := range tapCheck {
					fn(summaryKey)
				}

				// index may be out-dated, check store
				exists, err := to.Exists(summaryKey)
				ce(err)
				if exists {
					put(cont)
					return nil
				}

				// get summary
				var summary Summary
				ce(fetch(summaryKey, &summary))

				if len(summary.ReferedKeys) > 0 {
					// push refered first
					var n int64
					done := Proc(func() (err error) {
						defer he(&err)
						if atomic.AddInt64(&n, -1) != 0 {
							// not done
							return nil
						}
						// done
						ce(save(summaryKey, &summary))
						put(cont)
						return nil
					})
					var referedSummaryKeys []Key
					for _, key := range summary.ReferedKeys {
						var c int
						ce(selIndex(
							MatchEntry(IdxPairObjectSummary, key),
							Count(&c),
							Tap(func(_ Key, sKey Key) {
								referedSummaryKeys = append(
									referedSummaryKeys,
									sKey,
								)
							}),
						))
						if c == 0 {
							return we(fmt.Errorf(
								"no summary for %s",
								key,
							))
						}
					}
					n = int64(len(referedSummaryKeys))
					for _, sKey := range referedSummaryKeys {
						put(check(sKey, done))
					}

				} else {
					// no refered key
					ce(save(summaryKey, &summary))
					put(cont)
				}

				return
			}
		}

		// push
		iterKeys(func(summaryKey Key) (err error) {
			put(check(summaryKey, nil))
			return nil
		})

		ce(wait(false))
		ce(wait(true))

		return
	}
}

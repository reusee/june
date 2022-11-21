// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package entity

import (
	"fmt"
	"reflect"

	"github.com/reusee/june/index"
	"github.com/reusee/june/sys"
	"github.com/reusee/pr"
)

type Resave func(
	objs []any,
	options ...ResaveOption,
) error

type ResaveOption interface {
	IsResaveOption()
}

type SaveOptions []SaveOption

func (_ SaveOptions) IsResaveOption() {}

func (_ Def) Resave(
	sel index.SelectIndex,
	fetch Fetch,
	save Save,
	wt *pr.WaitTree,
	parallel sys.Parallel,
) Resave {

	return func(
		objs []any,
		options ...ResaveOption,
	) (err error) {
		defer he(&err)

		var tapKey []TapKey
		var saveOptions []SaveOption
		for _, option := range options {
			switch option := option.(type) {
			case TapKey:
				tapKey = append(tapKey, option)
			case SaveOptions:
				saveOptions = append(saveOptions, option...)
			default:
				panic(fmt.Errorf("unknown option: %T", option))
			}
		}

		wt := pr.NewWaitTree(wt)
		defer wt.Cancel()
		put, wait := pr.Consume(
			wt,
			int(parallel),
			func(_ int, v any) error {
				return v.(func() error)()
			},
		)
		defer func() {
			ce(wait(true))
		}()

		for _, obj := range objs {
			obj := obj
			objType := reflect.TypeOf(obj)
			ce(sel(
				MatchType(obj),
				TapKey(func(key Key) {

					put(func() (err error) {
						defer he(&err)
						for _, fn := range tapKey {
							fn(key)
						}
						ptr := reflect.New(objType)
						ce(fetch(key, ptr.Interface()))
						_, err = save(
							key.Namespace, ptr.Elem().Interface(),
							saveOptions...,
						)
						ce(err)
						return
					})

				}),
			))
		}

		return
	}
}

// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package entity

import (
	"context"
	"fmt"
	"reflect"

	"github.com/reusee/june/index"
	"github.com/reusee/june/sys"
	"github.com/reusee/pr2"
)

type Resave func(
	ctx context.Context,
	objs []any,
	options ...ResaveOption,
) error

type ResaveOption interface {
	IsResaveOption()
}

type SaveOptions []SaveOption

func (SaveOptions) IsResaveOption() {}

func (Def) Resave(
	sel index.SelectIndex,
	fetch Fetch,
	save Save,
	parallel sys.Parallel,
) Resave {

	return func(
		ctx context.Context,
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

		wg := pr2.NewWaitGroup(ctx)
		defer wg.Cancel()
		put, wait := pr2.Consume(
			wg,
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
				ctx,
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
							ctx,
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

// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package entity

import (
	"context"
	"fmt"
	"reflect"

	"github.com/reusee/june/opts"
	"github.com/reusee/june/store"
	"github.com/reusee/sb"
)

type CleanIndex func(
	ctx context.Context,
	options ...CleanIndexOption,
) error

type CleanIndexOption interface {
	IsCleanIndexOption()
}

func (Def) CleanIndex(
	deleteIndex DeleteIndex,
	store store.Store,
) CleanIndex {
	return func(
		ctx context.Context,
		options ...CleanIndexOption,
	) (err error) {
		defer he(&err)

		var deleteOptions []DeleteIndexOption

		var tapInvalid []opts.TapInvalidKey
		var tapKey []opts.TapKey
		for _, option := range options {
			if opt, ok := option.(DeleteIndexOption); ok {
				deleteOptions = append(deleteOptions, opt)
			}
			switch option := option.(type) {
			case opts.TapInvalidKey:
				tapInvalid = append(tapInvalid, option)
			case opts.TapKey:
				tapKey = append(tapKey, option)
			case DeleteIndexOption:
			default:
				panic(fmt.Errorf("bad option: %T", option))
			}
		}

		typeKey := reflect.TypeOf((*Key)(nil)).Elem()

		keyExists := make(map[Key]bool)
		ce(store.IterAllKeys(func(key Key) error {
			keyExists[key] = true
			return nil
		}))

		ce(deleteIndex(
			ctx,
			func(stream sb.Stream) (_ *IndexEntry, err error) {
				defer he(&err)

				hasInvalidKey := false
				unmarshal := func(ctx sb.Ctx, target reflect.Value, cont sb.Sink) sb.Sink {
					if target.Type().Elem() != typeKey {
						return sb.UnmarshalValue(ctx, target, cont)
					}

					return sb.UnmarshalValue(
						ctx, target,

						func(token *sb.Token) (next sb.Sink, err error) {
							defer he(&err)
							key := target.Elem().Interface().(Key)
							for _, fn := range tapKey {
								fn(key)
							}
							exists := keyExists[key]
							if !exists {
								for _, fn := range tapInvalid {
									fn(key)
								}
								hasInvalidKey = true
							}
							if cont != nil {
								return cont(token)
							}
							return nil, nil
						},
					)
				}

				var entry IndexEntry
				ce(sb.Copy(
					stream,
					unmarshal(
						sb.Ctx{
							Unmarshal: unmarshal,
						},
						reflect.ValueOf(&entry),
						nil,
					),
				))

				if hasInvalidKey {
					return &entry, nil
				}

				return nil, nil
			},
			deleteOptions...,
		))

		return nil
	}

}

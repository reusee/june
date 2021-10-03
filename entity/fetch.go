// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package entity

import (
	"fmt"
	"reflect"

	"github.com/reusee/e4"
	"github.com/reusee/june/key"
	"github.com/reusee/june/store"
	"github.com/reusee/sb"
)

type Fetch func(key any, targets ...any) error

func (_ Def) Fetch(
	store store.Store,
	newHashState key.NewHashState,
) Fetch {

	return func(arg any, targets ...any) (err error) {
		defer he(&err)

		var path []Key
		switch arg := arg.(type) {
		case Key:
			path = []Key{arg}
		case []Key:
			path = arg
		default:
			panic(fmt.Errorf("not key type: %T", arg))
		}

		ctx := sb.DefaultCtx.Strict()

		var sink sb.Sink
		if len(targets) == 1 {
			sink = sb.UnmarshalValue(ctx, reflect.ValueOf(targets[0]), nil)
		} else {
			var sinks []sb.Sink
			for _, target := range targets {
				sinks = append(sinks, sb.UnmarshalValue(ctx, reflect.ValueOf(target), nil))
			}
			sink = sb.AltSink(sinks...)
		}

		key := path[len(path)-1]
		checkTail := true
		for len(path) > 0 {
			var keyToCheck Key
			if checkTail {
				keyToCheck = path[len(path)-1]
				path = path[:len(path)-1]
			} else {
				keyToCheck = path[0]
				path = path[1:]
			}
			checkTail = !checkTail
			err := store.Read(keyToCheck, func(s sb.Stream) (err error) {
				defer he(&err)

				if keyToCheck == key {
					return sb.Copy(s, sink)
				}

				sub, err := sb.FindByHash(
					s, key.Hash[:], newHashState,
				)
				ce(err)
				return sb.Copy(sub, sink)

			})
			if is(err, ErrKeyNotFound) {
				continue
			}
			ce(err)
			return nil
		}

		return we.With(e4.With(key))(ErrKeyNotFound)
	}
}

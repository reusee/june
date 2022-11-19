// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package keyset

import (
	"context"
	"errors"

	"github.com/reusee/e5"
	"github.com/reusee/june/entity"
)

type Iter func(
	ctx context.Context,
	set Set,
	fn func(key Key) error,
) error

func (_ Def) Iter(
	fetch entity.Fetch,
) Iter {
	return func(
		ctx context.Context,
		set Set,
		fn func(key Key) error,
	) (err error) {
		return ce(
			set.iter(ctx, fetch, fn),
			e5.Ignore(Break),
		)
	}
}

var Break = errors.New("break")

func (s Set) iter(
	ctx context.Context,
	fetch entity.Fetch,
	fn func(key Key) error,
) error {
	for _, item := range s {
		if item.Key != nil {
			if err := fn(*item.Key); err != nil {
				return err
			}
		} else if item.Pack != nil {
			var set Set
			if err := fetch(ctx, item.Pack.Key, &set); err != nil {
				return err
			}
			if err := set.iter(ctx, fetch, fn); err != nil {
				return err
			}
		}
	}
	return nil
}

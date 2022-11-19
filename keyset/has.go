// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package keyset

import (
	"context"
	"sort"

	"github.com/reusee/june/entity"
)

type Has func(
	ctx context.Context,
	set Set,
	key Key,
) (
	ok bool,
	err error,
)

func (Def) Has(
	fetch entity.Fetch,
) Has {
	return func(
		ctx context.Context,
		set Set,
		key Key,
	) (
		ok bool,
		err error,
	) {
		return set.has(ctx, fetch, key)
	}
}

func (s Set) has(
	ctx context.Context,
	fetch entity.Fetch,
	key Key,
) (
	ok bool,
	err error,
) {
	defer he(&err)

	i := sort.Search(len(s), func(i int) bool {
		item := s[i]
		if item.Key != nil {
			return key.Compare(*item.Key) <= 0
		} else if item.Pack != nil {
			return key.Compare(item.Pack.Max) <= 0
		}
		panic("bad item")
	})

	if i >= len(s) {
		return false, nil
	}

	item := s[i]
	if item.Key != nil {
		return key == *item.Key, nil

	} else if item.Pack != nil {
		var set Set
		ce(fetch(ctx, item.Pack.Key, &set))
		return set.has(ctx, fetch, key)
	}

	panic("bad item")
}

// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storemonotree

import (
	"context"

	"github.com/reusee/june/store"
)

type Tree struct {
	upstream store.Store
	id       store.ID
}

type New func(
	ctx context.Context,
	upstream store.Store,
) (
	tree *Tree,
	err error,
)

func (Def) New() New {
	return func(
		ctx context.Context,
		upstream store.Store,
	) (
		tree *Tree,
		err error,
	) {
		defer he(&err)

		id, err := upstream.ID(ctx)
		ce(err)
		tree = &Tree{
			upstream: upstream,
			id:       "(monotree(" + id + "))",
		}

		return
	}
}

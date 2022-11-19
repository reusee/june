// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storepebble

import (
	"context"
	"testing"

	"github.com/reusee/e5"
	"github.com/reusee/june/index"
)

func TestIndex(
	t *testing.T,
	newStore New,
	test index.TestIndex,
) {
	defer he(nil, e5.TestingFatal(t))
	ctx := context.Background()

	dir := t.TempDir()

	s, err := newStore(ctx, nil, dir)
	ce(err)
	with := func(fn func(index.IndexManager)) {
		fn(s)
	}
	test(ctx, with, t)
}

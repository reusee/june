// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storemem

import (
	"context"
	"testing"

	"github.com/reusee/june/index"
	"github.com/reusee/june/storekv"
)

func TestStore(
	t *testing.T,
	test storekv.TestKV,
	newStore New,
) {
	ctx := context.Background()
	with := func(fn func(storekv.KV, string)) {
		m := newStore()
		fn(m, "foo")
	}
	test(ctx, t, with)
}

func TestIndex(
	t *testing.T,
	test index.TestIndex,
	newManager New,
) {
	ctx := context.Background()
	manager := newManager()
	with := func(fn func(index.IndexManager)) {
		fn(manager)
	}
	test(ctx, with, t)
}

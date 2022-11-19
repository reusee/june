// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storepebble

import (
	"context"
	"os"
	"testing"

	"github.com/reusee/e5"
	"github.com/reusee/june/index"
	"github.com/reusee/june/storekv"
)

func TestBatchKV(
	t *testing.T,
	test storekv.TestKV,
	newStore New,
	newBatch NewBatch,
) {
	defer he(nil, e5.TestingFatal(t))
	ctx := context.Background()
	with := func(fn func(storekv.KV, string)) {
		dir, err := os.MkdirTemp(t.TempDir(), "")
		ce(err)
		s, err := newStore(ctx, nil, dir)
		ce(err)
		batch, err := newBatch(ctx, s)
		ce(err)
		fn(batch, "foo")
	}
	test(ctx, t, with)
}

func TestBatchIndex(
	t *testing.T,
	newStore New,
	newBatch NewBatch,
	test index.TestIndex,
) {
	defer he(nil, e5.TestingFatal(t))
	ctx := context.Background()
	dir, err := os.MkdirTemp(t.TempDir(), "")
	ce(err)
	s, err := newStore(ctx, nil, dir)
	ce(err)
	batch, err := newBatch(ctx, s)
	ce(err)
	with := func(fn func(index.IndexManager)) {
		fn(batch)
	}
	test(ctx, with, t)
}

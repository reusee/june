// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storepebble

import (
	"os"
	"testing"

	"github.com/reusee/dscope"
	"github.com/reusee/e5"
	"github.com/reusee/june/index"
	"github.com/reusee/june/storekv"
	"github.com/reusee/pr"
)

func TestBatchKV(
	t *testing.T,
	wt *pr.WaitTree,
	test storekv.TestKV,
	scope dscope.Scope,
	newStore New,
	newBatch NewBatch,
) {
	defer he(nil, e5.TestingFatal(t))
	with := func(fn func(storekv.KV, string)) {
		dir, err := os.MkdirTemp(t.TempDir(), "")
		ce(err)
		s, err := newStore(wt, nil, dir)
		ce(err)
		batch, err := newBatch(wt, s)
		ce(err)
		fn(batch, "foo")
	}
	test(t, with)
}

func TestBatchIndex(
	t *testing.T,
	wt *pr.WaitTree,
	newStore New,
	newBatch NewBatch,
	test index.TestIndex,
) {
	defer he(nil, e5.TestingFatal(t))
	dir, err := os.MkdirTemp(t.TempDir(), "")
	ce(err)
	s, err := newStore(wt, nil, dir)
	ce(err)
	batch, err := newBatch(wt, s)
	ce(err)
	with := func(fn func(index.IndexManager)) {
		fn(batch)
	}
	test(with, t)
}

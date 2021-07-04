// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storepebble

import (
	"os"
	"testing"

	"github.com/reusee/dscope"
	"github.com/reusee/e4"
	"github.com/reusee/ling/v2/index"
	"github.com/reusee/ling/v2/storekv"
)

func TestBatchKV(
	t *testing.T,
	test storekv.TestKV,
	scope dscope.Scope,
	newStore New,
	newBatch NewBatch,
) {
	defer he(nil, e4.TestingFatal(t))
	with := func(fn func(storekv.KV, string)) {
		dir, err := os.MkdirTemp(t.TempDir(), "")
		ce(err)
		s, err := newStore(nil, dir)
		ce(err)
		defer s.Close()
		batch, err := newBatch(s)
		ce(err)
		defer batch.Close()
		fn(batch, "foo")
	}
	test(t, with)
}

func TestBatchIndex(
	t *testing.T,
	newStore New,
	newBatch NewBatch,
	test index.TestIndex,
) {
	defer he(nil, e4.TestingFatal(t))
	dir, err := os.MkdirTemp(t.TempDir(), "")
	ce(err)
	s, err := newStore(nil, dir)
	ce(err)
	defer s.Close()
	batch, err := newBatch(s)
	ce(err)
	defer batch.Close()
	with := func(fn func(index.IndexManager)) {
		fn(batch)
	}
	test(with, t)
}

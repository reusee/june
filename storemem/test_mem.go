// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storemem

import (
	"testing"

	"github.com/reusee/june/index"
	"github.com/reusee/june/storekv"
	"github.com/reusee/pr"
)

func TestStore(
	t *testing.T,
	wt *pr.WaitTree,
	test storekv.TestKV,
	newStore New,
) {
	with := func(fn func(storekv.KV, string)) {
		m := newStore(wt)
		fn(m, "foo")
	}
	test(t, with)
}

func TestIndex(
	t *testing.T,
	wt *pr.WaitTree,
	test index.TestIndex,
	newManager New,
) {
	manager := newManager(wt)
	with := func(fn func(index.IndexManager)) {
		fn(manager)
	}
	test(with, t)
}

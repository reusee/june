// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storemem

import (
	"testing"

	"github.com/reusee/june/index"
	"github.com/reusee/june/storekv"
	"github.com/reusee/pr2"
)

func TestStore(
	t *testing.T,
	test storekv.TestKV,
	newStore New,
	wg *pr2.WaitGroup,
) {
	with := func(fn func(storekv.KV, string)) {
		m := newStore(wg)
		fn(m, "foo")
	}
	test(t, with)
}

func TestIndex(
	t *testing.T,
	test index.TestIndex,
	newManager New,
	wg *pr2.WaitGroup,
) {
	manager := newManager(wg)
	with := func(fn func(index.IndexManager)) {
		fn(manager)
	}
	test(with, t)
}

// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storepebble

import (
	"testing"

	"github.com/reusee/e4"
	"github.com/reusee/june/index"
	"github.com/reusee/pr"
)

func TestIndex(
	t *testing.T,
	wt *pr.WaitTree,
	newStore New,
	test index.TestIndex,
) {
	defer he(nil, e4.TestingFatal(t))

	dir := t.TempDir()

	s, err := newStore(wt, nil, dir)
	ce(err)
	with := func(fn func(index.IndexManager)) {
		fn(s)
	}
	test(with, t)
}

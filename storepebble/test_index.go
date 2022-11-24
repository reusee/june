// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storepebble

import (
	"testing"

	"github.com/reusee/e5"
	"github.com/reusee/june/index"
	"github.com/reusee/pr2"
)

func TestIndex(
	t *testing.T,
	newStore New,
	test index.TestIndex,
	wg *pr2.WaitGroup,
) {
	defer he(nil, e5.TestingFatal(t))

	dir := t.TempDir()

	s, err := newStore(wg, nil, dir)
	ce(err)
	with := func(fn func(index.IndexManager)) {
		fn(s)
	}
	test(with, t)
}

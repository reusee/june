// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storepebble

import (
	"testing"

	"github.com/reusee/e4"
	"github.com/reusee/june/index"
)

func TestIndex(
	t *testing.T,
	newStore New,
	test index.TestIndex,
) {
	defer he(nil, e4.TestingFatal(t))

	dir := t.TempDir()

	s, err := newStore(nil, dir)
	ce(err)
	defer s.Close()
	with := func(fn func(index.IndexManager)) {
		fn(s)
	}
	test(with, t)
}

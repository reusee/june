// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storemem

import (
	"testing"

	"github.com/reusee/ling/v2/index"
	"github.com/reusee/ling/v2/storekv"
)

func TestStore(
	t *testing.T,
	test storekv.TestKV,
	newStore New,
) {
	with := func(fn func(storekv.KV, string)) {
		m := newStore()
		fn(m, "foo")
	}
	test(t, with)
}

func TestIndex(
	t *testing.T,
	test index.TestIndex,
	newManager New,
) {
	manager := newManager()
	with := func(fn func(index.IndexManager)) {
		fn(manager)
	}
	test(with, t)
}

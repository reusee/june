// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storepebble

import (
	"os"
	"testing"

	"github.com/reusee/e5"
	"github.com/reusee/june/storekv"
	"github.com/reusee/pr2"
)

func TestKV(
	t *testing.T,
	test storekv.TestKV,
	newStore New,
	wg *pr2.WaitGroup,
) {
	defer he(nil, e5.TestingFatal(t))
	with := func(fn func(storekv.KV, string)) {
		dir, err := os.MkdirTemp(t.TempDir(), "")
		ce(err)
		s, err := newStore(wg, nil, dir)
		ce(err)
		fn(s, "foo")
	}
	test(wg, t, with)
}

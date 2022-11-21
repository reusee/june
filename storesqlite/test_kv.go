// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storesqlite

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/reusee/dscope"
	"github.com/reusee/e5"
	"github.com/reusee/june/storekv"
	"github.com/reusee/pr"
)

func TestKV(
	t *testing.T,
	wt *pr.WaitTree,
	test storekv.TestKV,
	scope dscope.Scope,
	newStore New,
) {
	defer he(nil, e5.TestingFatal(t))
	with := func(fn func(storekv.KV, string)) {
		dir, err := os.MkdirTemp(t.TempDir(), "")
		ce(err)
		s, err := newStore(
			wt,
			filepath.Join(dir, fmt.Sprintf("%d", rand.Int63())),
		)
		ce(err)
		fn(s, "foo")
	}
	test(t, with)
}

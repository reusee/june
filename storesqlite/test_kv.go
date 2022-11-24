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
	t.Skip() //TODO

	defer he(nil, e5.TestingFatal(t))
	with := func(fn func(storekv.KV, string)) {
		dir, err := os.MkdirTemp(t.TempDir(), "")
		ce(err)
		s, err := newStore(
			wg,
			filepath.Join(dir, fmt.Sprintf("%d", rand.Int63())),
		)
		ce(err)
		fn(s, "foo")
	}
	test(wg, t, with)
}

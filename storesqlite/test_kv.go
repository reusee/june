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
	"github.com/reusee/e4"
	"github.com/reusee/june/storekv"
)

func TestKV(
	t *testing.T,
	test storekv.TestKV,
	scope dscope.Scope,
	newStore New,
) {
	defer he(nil, e4.TestingFatal(t))
	with := func(fn func(storekv.KV, string)) {
		dir, err := os.MkdirTemp(t.TempDir(), "")
		ce(err)
		s, err := newStore(
			filepath.Join(dir, fmt.Sprintf("%d", rand.Int63())),
		)
		ce(err)
		fn(s, "foo")
	}
	test(t, with)
}

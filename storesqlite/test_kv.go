// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storesqlite

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/reusee/e5"
	"github.com/reusee/june/storekv"
)

func TestKV(
	t *testing.T,
	test storekv.TestKV,
	newStore New,
) {
	defer he(nil, e5.TestingFatal(t))
	ctx := context.Background()
	with := func(fn func(storekv.KV, string)) {
		dir, err := os.MkdirTemp(t.TempDir(), "")
		ce(err)
		s, err := newStore(
			ctx,
			filepath.Join(dir, fmt.Sprintf("%d", rand.Int63())),
		)
		ce(err)
		fn(s, "foo")
	}
	test(ctx, t, with)
}

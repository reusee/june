// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storedisk

import (
	"context"
	"os"
	"testing"

	"github.com/reusee/e5"
	"github.com/reusee/june/storekv"
)

func TestStore(
	t *testing.T,
	test storekv.TestKV,
	newStore New,
) {
	defer he(nil, e5.TestingFatal(t))
	ctx := context.Background()
	with := func(fn func(storekv.KV, string)) {
		dir, err := os.MkdirTemp(t.TempDir(), "")
		ce(err)
		s, err := newStore(dir)
		ce(err)
		fn(s, "foo")
	}
	test(ctx, t, with)
}

func TestStoreSoftDelete(
	t *testing.T,
	test storekv.TestKV,
	newStore New,
) {
	defer he(nil, e5.TestingFatal(t))
	ctx := context.Background()
	with := func(fn func(storekv.KV, string)) {
		dir, err := os.MkdirTemp(t.TempDir(), "")
		ce(err)
		s, err := newStore(dir, SoftDelete(true))
		ce(err)
		fn(s, "foo")
	}
	test(ctx, t, with)
}

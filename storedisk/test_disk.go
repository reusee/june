// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storedisk

import (
	"os"
	"testing"

	"github.com/reusee/e4"
	"github.com/reusee/june/storekv"
)

func TestStore(
	t *testing.T,
	test storekv.TestKV,
	newStore New,
) {
	defer he(nil, e4.TestingFatal(t))
	with := func(fn func(storekv.KV, string)) {
		dir, err := os.MkdirTemp(t.TempDir(), "")
		ce(err)
		s, err := newStore(dir)
		ce(err)
		defer s.Close()
		fn(s, "foo")
	}
	test(t, with)
}

func TestStoreSoftDelete(
	t *testing.T,
	test storekv.TestKV,
	newStore New,
) {
	defer he(nil, e4.TestingFatal(t))
	with := func(fn func(storekv.KV, string)) {
		dir, err := os.MkdirTemp(t.TempDir(), "")
		ce(err)
		s, err := newStore(dir, SoftDelete(true))
		ce(err)
		defer s.Close()
		fn(s, "foo")
	}
	test(t, with)
}

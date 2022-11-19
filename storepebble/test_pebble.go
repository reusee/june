// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storepebble

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/reusee/e5"
	"github.com/reusee/june/index"
	"github.com/reusee/june/storekv"
	"github.com/reusee/sb"
)

func TestMixedIndex(
	t *testing.T,
	newStore New,
	testIndex index.TestIndex,
) {
	defer he(nil, e5.TestingFatal(t))
	ctx := context.Background()

	withIndex := func(fn func(index.IndexManager)) {
		dir := t.TempDir()
		s, err := newStore(ctx, nil, dir)
		ce(err)

		buf := new(bytes.Buffer)
		if err := sb.Copy(
			sb.Marshal(42),
			sb.Encode(buf),
		); err != nil {
			t.Fatal(err)
		}
		err = s.KeyPut(
			ctx,
			"foo",
			buf,
		)
		ce(err)

		fn(s)
	}
	testIndex(ctx, withIndex, t)

}

func TestMixedKV(
	t *testing.T,
	newStore New,
	testKV storekv.TestKV,
) {
	ctx := context.Background()

	withKV := func(fn func(storekv.KV, string)) {
		dir, err := os.MkdirTemp(t.TempDir(), "")
		ce(err)
		s, err := newStore(ctx, nil, dir)
		ce(err)

		entry := index.NewEntry(TestingIndex, 42)
		entry.Key = &Key{}

		index, err := s.IndexFor("foo")
		ce(err)
		err = index.Save(ctx, entry)
		ce(err)

		fn(s, "foo")
	}
	testKV(ctx, t, withKV)

}

type testingIndex struct {
	Int int
}

var TestingIndex = testingIndex{}

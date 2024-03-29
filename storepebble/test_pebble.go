// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storepebble

import (
	"bytes"
	"os"
	"testing"

	"github.com/reusee/e5"
	"github.com/reusee/june/index"
	"github.com/reusee/june/storekv"
	"github.com/reusee/pr2"
	"github.com/reusee/sb"
)

func TestMixedIndex(
	t *testing.T,
	newStore New,
	testIndex index.TestIndex,
	wg *pr2.WaitGroup,
) {
	defer he(nil, e5.TestingFatal(t))

	withIndex := func(fn func(index.IndexManager)) {
		dir := t.TempDir()
		s, err := newStore(wg, nil, dir)
		ce(err)

		buf := new(bytes.Buffer)
		if err := sb.Copy(
			sb.Marshal(42),
			sb.Encode(buf),
		); err != nil {
			t.Fatal(err)
		}
		err = s.KeyPut(
			"foo",
			buf,
		)
		ce(err)

		fn(s)
	}
	testIndex(withIndex, t)

}

func TestMixedKV(
	t *testing.T,
	newStore New,
	testKV storekv.TestKV,
	wg *pr2.WaitGroup,
) {

	withKV := func(fn func(storekv.KV, string)) {
		dir, err := os.MkdirTemp(t.TempDir(), "")
		ce(err)
		s, err := newStore(wg, nil, dir)
		ce(err)

		entry := index.NewEntry(TestingIndex, 42)
		entry.Key = &Key{}

		index, err := s.IndexFor("foo")
		ce(err)
		err = index.Save(entry)
		ce(err)

		fn(s, "foo")
	}
	testKV(wg, t, withKV)

}

type testingIndex struct {
	Int int
}

var TestingIndex = testingIndex{}

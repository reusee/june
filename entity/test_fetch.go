// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package entity

import (
	"testing"

	"github.com/reusee/e5"
	"github.com/reusee/pr2"
)

type testFetch struct {
	Foo testFetch1
}

type testFetch1 int

var _ HasIndex = testFetch1(0)

func (t testFetch1) EntityIndexes() (IndexSet, int64, error) {
	return IndexSet{
		IdxName(Name("foo")),
	}, 1, nil
}

func TestFetch(
	t *testing.T,
	fetch Fetch,
	save SaveEntity,
	wg *pr2.WaitGroup,
) {
	defer he(nil, e5.TestingFatal(t))

	summary, err := save(wg, testFetch{
		Foo: testFetch1(42),
	})
	ce(err)

	var f testFetch
	ce(fetch(summary.Key, &f))
	if f.Foo != 42 {
		t.Fatal()
	}

	var f1 testFetch1
	ce(fetch([]Key{
		summary.Key,
		summary.Subs[0].Key,
	}, &f1))
	if f1 != 42 {
		t.Fatal()
	}

}

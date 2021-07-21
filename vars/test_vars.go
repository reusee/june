// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package vars

import (
	"testing"

	"github.com/reusee/e4"
	"github.com/reusee/pr"
)

func TestVars(
	t *testing.T,
	scope Scope,
	wt *pr.WaitTree,
) {
	defer he(nil, e4.TestingFatal(t))

	scope.Sub(func() VarsSpec {
		return func() (string, *pr.WaitTree) {
			return t.TempDir(), wt
		}
	}).Call(func(
		get Get,
		set Set,
		store *VarsStore,
	) {
		defer store.db.Close()

		var v any
		err := get("foo", &v)
		if e := new(NotFound); !as(err, &e) {
			t.Fatalf("got %v", err)
		}

		err = set("foo", 42)
		ce(err)
		err = get("foo", &v)
		ce(err)
		if v != 42 {
			t.Fatal()
		}

		err = get("bar", &v)
		if e := new(NotFound); !as(err, &e) {
			t.Fatalf("got %v", err)
		}
		if v != 42 { // not changed
			t.Fatal()
		}

	})

}

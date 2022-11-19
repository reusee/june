// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package vars

import (
	"context"
	"testing"

	"github.com/reusee/e5"
)

func TestVars(
	t *testing.T,
	scope Scope,
) {
	defer he(nil, e5.TestingFatal(t))
	ctx := context.Background()

	scope.Fork(func() VarsSpec {
		return func() (string, context.Context) {
			return t.TempDir(), ctx
		}
	}).Call(func(
		get Get,
		set Set,
		store *VarsStore,
	) {
		defer store.db.Close()

		var v any
		err := get(ctx, "foo", &v)
		if e := new(NotFound); !as(err, &e) {
			t.Fatalf("got %v", err)
		}

		err = set(ctx, "foo", 42)
		ce(err)
		err = get(ctx, "foo", &v)
		ce(err)
		if v != 42 {
			t.Fatal()
		}

		err = get(ctx, "bar", &v)
		if e := new(NotFound); !as(err, &e) {
			t.Fatalf("got %v", err)
		}
		if v != 42 { // not changed
			t.Fatal()
		}

	})

}

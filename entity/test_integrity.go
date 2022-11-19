// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package entity

import (
	"context"
	"testing"

	"github.com/reusee/e5"
)

func TestIntegrity(
	t *testing.T,
	saveEntity SaveEntity,
	store Store,
) {
	defer he(nil, e5.TestingFatal(t))
	ctx := context.Background()

	type Foo int
	type Bar struct {
		Key Key
	}
	s, err := saveEntity(ctx, Foo(42))
	ce(err)
	s2, err := saveEntity(ctx, Bar{
		Key: s.Key,
	})
	ce(err)

	ce(s2.checkRef(ctx, store))

	ce(store.Delete(ctx, []Key{s.Key}))
	err = s2.checkRef(ctx, store)
	if !is(err, ErrKeyNotFound) {
		t.Fatal()
	}

}

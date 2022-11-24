// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package entity

import (
	"testing"

	"github.com/reusee/e5"
	"github.com/reusee/pr2"
)

func TestIntegrity(
	t *testing.T,
	saveEntity SaveEntity,
	store Store,
	wg *pr2.WaitGroup,
) {
	defer he(nil, e5.TestingFatal(t))

	type Foo int
	type Bar struct {
		Key Key
	}
	s, err := saveEntity(wg, Foo(42))
	ce(err)
	s2, err := saveEntity(wg, Bar{
		Key: s.Key,
	})
	ce(err)

	ce(s2.checkRef(store))

	ce(store.Delete([]Key{s.Key}))
	err = s2.checkRef(store)
	if !is(err, ErrKeyNotFound) {
		t.Fatal()
	}

}

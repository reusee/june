// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package keyset

import (
	"crypto/rand"
	"testing"
)

func TestSet(
	t *testing.T,
	scope Scope,
) {

	threshold := Threshold(4)
	scope.Sub(&threshold).Call(func(
		add Add,
	) {

		var keys []Key
		for i := 0; i < 2048; i++ {
			var key Key
			key.Namespace[0] = 'a' + byte(i%32)
			_, err := rand.Read(key.Hash[:])
			ce(err)
			keys = append(keys, key)
		}
		set, err := add(Set{}, keys...)
		ce(err)

		if len(set) == 0 {
			t.Fatal()
		}

	})

}

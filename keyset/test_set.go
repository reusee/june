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

	threshold := PackThreshold(4)
	scope.Fork(&threshold).Call(func(
		add Add,
		iter Iter,
		has Has,
		pack PackSet,
		del Delete,
	) {

		// add and pack
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
		set, err = pack(set)
		ce(err)

		if len(set) == 0 {
			t.Fatal()
		}

		// iter
		var ks []Key
		ce(iter(set, func(key Key) error {
			ks = append(ks, key)
			return nil
		}))
		if len(ks) != len(keys) {
			t.Fatal()
		}

		m := make(map[Key]bool)
		for _, key := range keys {
			m[key] = true
		}
		for _, key := range ks {
			if _, ok := m[key]; !ok {
				t.Fatal()
			}
		}

		m = make(map[Key]bool)
		for _, key := range ks {
			m[key] = true
		}
		for _, key := range keys {
			if _, ok := m[key]; !ok {
				t.Fatal()
			}
		}

		for i, key := range ks {
			if i == 0 {
				continue
			}
			last := ks[i-1]
			if key.Compare(last) != 1 {
				t.Fatal()
			}
		}

		for _, key := range keys {
			ok, err := has(set, key)
			ce(err)
			if !ok {
				t.Fatal()
			}
		}

		// delete
		for i, key := range keys {
			set, err = del(set, key)
			ce(err)
			if i%128 == 0 {
				set, err = pack(set)
				ce(err)
			}
		}
		if len(set) != 0 {
			t.Fatal()
		}

	})

}

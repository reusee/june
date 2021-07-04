// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package key

import "testing"

func TestKeyPath(t *testing.T) {
	p1 := KeyPath{
		{Hash: Hash{'a'}},
		{Hash: Hash{'b'}},
	}
	p2 := KeyPath{
		{Hash: Hash{'a'}},
		{Hash: Hash{'c'}},
		{Hash: Hash{'b'}},
	}
	if !p1.Same(p2) {
		t.Fatal()
	}

	p1 = KeyPath{
		{Hash: Hash{'a'}},
	}
	p2 = KeyPath{
		{Hash: Hash{'a'}},
		{Hash: Hash{'c'}},
		{Hash: Hash{'b'}},
	}
	if p1.Same(p2) {
		t.Fatal()
	}

	p1 = KeyPath{}
	p2 = KeyPath{}
	if !p1.Same(p2) {
		t.Fatal()
	}

	p1 = KeyPath{
		{Hash: Hash{'a'}},
	}
	p2 = KeyPath{}
	if p1.Same(p2) {
		t.Fatal()
	}
}

// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package naming

import (
	"reflect"
	"testing"

	"github.com/reusee/e4"
)

func TestTypeName(t *testing.T) {
	defer he(nil, e4.TestingFatal(t))

	type S string
	name := Type(reflect.TypeOf((*S)(nil)).Elem())
	if name != "naming.S" {
		t.Fatalf("got %s\n", name)
	}

	name = Type(reflect.TypeOf((**S)(nil)).Elem())
	if name != "*naming.S" {
		t.Fatal()
	}

	name = Type(reflect.TypeOf((***S)(nil)).Elem())
	if name != "**naming.S" {
		t.Fatal()
	}

	name = Type(reflect.TypeOf((*int)(nil)).Elem())
	if name != "int" {
		t.Fatal()
	}

	type Foo int
	type Bar = Foo
	nameFoo := Type(reflect.TypeOf((*Foo)(nil)).Elem())
	nameBar := Type(reflect.TypeOf((*Bar)(nil)).Elem())
	if nameFoo != nameBar {
		t.Fatal()
	}

}

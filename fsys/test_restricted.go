// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package fsys

import (
	"testing"

	"github.com/reusee/e5"
)

func TestRestrictedPath(
	t *testing.T,
	set SetRestrictedPath,
	is IsRestrictedPath,
) {
	defer he(nil, e5.TestingFatal(t))

	// root
	ok, err := is("/")
	ce(err)
	if ok {
		t.Fatal()
	}

	// foo not set
	ok, err = is("foo")
	ce(err)
	if ok {
		t.Fatal()
	}

	// set foo
	err = set("foo")
	ce(err)
	ok, err = is("foo")
	ce(err)
	if !ok {
		t.Fatal()
	}

	// sub dir of foo
	ok, err = is("foo/bar")
	ce(err)
	if !ok {
		t.Fatal()
	}

	// bar not set
	ok, err = is("bar")
	ce(err)
	if ok {
		t.Fatal()
	}

	// set bar/bar/bar
	err = set("bar/bar/bar")
	ce(err)
	ok, err = is("bar/bar/bar")
	ce(err)
	if !ok {
		t.Fatal()
	}

	// parent not set
	ok, err = is("bar/bar")
	ce(err)
	if ok {
		t.Fatal()
	}

	// sub dir
	ok, err = is("bar/bar/bar/baz")
	ce(err)
	if !ok {
		t.Fatal()
	}

}

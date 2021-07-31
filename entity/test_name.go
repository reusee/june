// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package entity

import (
	"strings"
	"testing"
	"time"

	"github.com/reusee/june/naming"
)

func TestNewName(
	t *testing.T,
	newName NewName,
	machineName naming.MachineName,
) {
	name := newName("foo")
	if !name.Valid() {
		t.Fatal()
	}
	str := string(name)
	if !strings.HasPrefix(str, "foo/") {
		t.Fatal()
	}
	if !strings.Contains(str, string(machineName)) {
		t.Fatal()
	}
	if !strings.Contains(str, time.Now().Format("20060102")) {
		t.Fatal()
	}
}

// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package key

import (
	"testing"

	"github.com/reusee/e5"
)

func TestHashValue(t *testing.T) {
	defer he(nil, e5.TestingFatal(t))
	h, err := HashValue(42)
	ce(err)
	if h.String() != "151a3a0b4c88483512fc484d0badfedf80013ebb18df498bbee89ac5b69d7222" {
		t.Fatal()
	}
}

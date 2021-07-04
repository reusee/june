// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package sys

import "testing"

type Testing bool

func (_ Def) Testing() Testing {
	return false
}

func TestTesting(
	t *testing.T,
	testing Testing,
) {
	if !testing {
		t.Fatal()
	}
}

// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package config

import (
	"testing"

	"github.com/reusee/e4"
)

func TestConfig(
	t *testing.T,
	scope Scope,
) {
	defer he(nil, e4.TestingFatal(t))

	scope.Sub(func() UserConfig {
		return UserConfig(`
Foo = 42
Bar = "bar"
    `)
	}).Call(func(
		get GetConfig,
	) {

		var foo struct {
			Bar string
			Foo int
		}
		err := get(&foo)
		ce(err)
		if foo.Foo != 42 {
			t.Fatal()
		}
		if foo.Bar != "bar" {
			t.Fatal()
		}

	})
}

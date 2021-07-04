// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package naming

import (
	"os"
)

type MachineName string

func (_ Def) DefaultMachineName() MachineName {
	name, _ := os.Hostname()
	if name == "" {
		name = DefaultMachineName
	}
	return MachineName(name)
}

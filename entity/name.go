// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package entity

import (
	"crypto/rand"
	"fmt"
	"path"

	"github.com/reusee/ling/v2/clock"
	"github.com/reusee/ling/v2/naming"
)

// Name

type Name string

func (n Name) Valid() bool {
	return n != ""
}

// NewName

type NewName func(prefix string) Name

func (_ Def) NewName(
	machineName naming.MachineName,
	now clock.Now,
	index Index,
) NewName {

	return func(prefix string) Name {
		bs := make([]byte, 2)
		for {
			if _, err := rand.Read(bs[:]); err != nil {
				panic(err)
			}
			t := now()
			name := Name(path.Join(
				prefix,
				fmt.Sprintf(
					"%s-%s%03d-%x",
					machineName,
					t.Format("20060102150405"),
					t.Nanosecond()/1000/1000,
					bs,
				),
			))

			// check dup
			var n int
			if err := Select(
				index,
				MatchEntry(IdxName, name),
				Count(&n),
			); err != nil {
				panic(err)
			}
			if n > 0 {
				// duplicated
				continue
			}

			return name
		}
	}
}

// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package index

import (
	"github.com/reusee/pr"
)

type SelectIndex func(
	args ...SelectOption,
) error

func (_ Def) SelectIndex(
	index Index,
	wt *pr.WaitTree,
) SelectIndex {
	return func(args ...SelectOption) (err error) {
		defer he(&err)
		args = append(args, WithCtx{wt.Ctx})
		return Select(index, args...)
	}
}

// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package index

type SelectIndex func(
	args ...SelectOption,
) error

func (_ Def) SelectIndex(
	index Index,
) SelectIndex {
	return func(args ...SelectOption) (err error) {
		defer he(&err)
		return Select(index, args...)
	}
}

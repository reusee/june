// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package index

import "context"

type SelectIndex func(
	ctx context.Context,
	args ...SelectOption,
) error

func (Def) SelectIndex(
	index Index,
) SelectIndex {
	return func(ctx context.Context, args ...SelectOption) (err error) {
		defer he(&err)
		args = append(args, WithCtx{ctx})
		return Select(ctx, index, args...)
	}
}

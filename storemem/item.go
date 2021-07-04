// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storemem

import (
	"github.com/google/btree"
	"github.com/reusee/sb"
)

type Item struct {
	sb.Tokens
}

var _ btree.Item = Item{}

func (i Item) Less(than btree.Item) bool {
	return sb.MustCompare(
		i.Iter(),
		than.(Item).Iter(),
	) < 0
}

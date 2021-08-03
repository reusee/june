// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package keyset

type Set []SetItem

type SetItem struct {
	Key  *Key
	Pack *Pack
}

type Pack struct {
	Min    Key
	Max    Key
	Height int
	Key    Key // key of KeySet
}

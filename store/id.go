// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package store

func (Def) StoreID(
	store Store,
) ID {
	id, err := store.ID()
	ce(err)
	return id
}

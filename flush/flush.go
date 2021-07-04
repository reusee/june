// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package flush

import (
	"github.com/reusee/ling/v2/index"
	"github.com/reusee/ling/v2/store"
)

type Flush func() error

func (_ Def) Flush(
	store store.Store,
	indexManager index.IndexManager,
) Flush {
	return func() error {
		if err := store.Sync(); err != nil {
			return err
		}
		if err := indexManager.Sync(); err != nil {
			return err
		}
		return nil
	}
}

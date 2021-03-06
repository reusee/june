// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package tx

import (
	"github.com/reusee/dscope"
	"github.com/reusee/e4"
	"github.com/reusee/june/store"
	"github.com/reusee/june/storekv"
	"github.com/reusee/june/storepebble"
	"github.com/reusee/pr"
)

type KVToStore func(kv storekv.KV) (store.Store, error)

type PebbleTx func(
	wt *pr.WaitTree,
	fn any,
) error

func UsePebbleTx(
	kvToStore KVToStore,
	peb *storepebble.Store,
	newBatch storepebble.NewBatch,
	scope dscope.Scope,
) PebbleTx {

	return func(wt *pr.WaitTree, fn any) (err error) {
		defer he(&err)

		batch, err := newBatch(wt, peb)
		ce(err)
		defer he(&err, e4.WrapFunc(func(err error) error {
			if e := batch.Abort(); e != nil {
				return e4.MakeErr(e, err)
			}
			return err
		}))

		scope.Fork(
			func() Store {
				kv, err := kvToStore(batch)
				ce(err)
				return kv
			},
			func() IndexManager {
				return batch
			},
		).Call(fn)

		ce(batch.Commit())

		return
	}
}

// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package scopeproxy

import (
	"github.com/reusee/june/entity"
	"github.com/reusee/june/key"
	"github.com/reusee/june/tx"
	"github.com/reusee/pr"
)

func EntitySave(
	tx tx.PebbleTx,
	wt *pr.WaitTree,
) entity.Save {
	return func(
		ns key.Namespace,
		value any,
		options ...entity.SaveOption,
	) (
		summary *entity.Summary,
		err error,
	) {
		if e := tx(wt, func(
			save entity.Save,
		) {
			summary, err = save(ns, value, options...)
		}); e != nil {
			err = e
		}
		return
	}
}

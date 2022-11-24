// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package scopeproxy

import (
	"context"

	"github.com/reusee/june/entity"
	"github.com/reusee/june/key"
	"github.com/reusee/june/tx"
)

func EntitySave(
	tx tx.PebbleTx,
) entity.Save {
	return func(
		ctx context.Context,
		ns key.Namespace,
		value any,
		options ...entity.SaveOption,
	) (
		summary *entity.Summary,
		err error,
	) {
		if e := tx(ctx, func(
			save entity.Save,
		) {
			summary, err = save(ctx, ns, value, options...)
		}); e != nil {
			err = e
		}
		return
	}
}

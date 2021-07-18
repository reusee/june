// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package index

import (
	"reflect"

	"github.com/reusee/sb"
)

type StoreIndex struct {
	Value sb.Stream
	ID    StoreID
}

var _ sb.SBMarshaler = StoreIndex{}

func (s StoreIndex) MarshalSB(ctx sb.Ctx, cont sb.Proc) sb.Proc {
	return ctx.Marshal(
		ctx,
		reflect.ValueOf(sb.Tuple{s.ID, s.Value}),
		cont,
	)
}

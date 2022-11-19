// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package index

import (
	"context"
	"errors"
	"io"

	"github.com/reusee/sb"
)

type IndexManager interface {
	Name() string
	IndexFor(StoreID) (Index, error)
}

type SaveOption interface {
	IsSaveOption()
}

func (TapEntry) IsSaveOption() {}

type Index interface {
	Name(ctx context.Context) (string, error)
	Iter(
		ctx context.Context,
		lower *sb.Tokens, // inclusive in any order
		upper *sb.Tokens, // exclusive in any order
		order Order,
	) (Src, io.Closer, error)
	Save(ctx context.Context, entry Entry, options ...SaveOption) error
	Delete(ctx context.Context, entry Entry) error
}

var ErrInvalidEntry = errors.New("invalid entry")

type Order uint8

const (
	Asc Order = iota
	Desc
)

func (Order) IsSelectOption() {}

func (Order) IsIterOption() {}

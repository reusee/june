// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package index

import (
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

func (_ TapEntry) IsSaveOption() {}

type Index interface {
	Name() string
	Iter(
		lower *sb.Tokens, // inclusive in any order
		upper *sb.Tokens, // exclusive in any order
		order Order,
	) (Src, io.Closer, error)
	Save(entry Entry, options ...SaveOption) error
	Delete(entry Entry) error
}

var ErrInvalidEntry = errors.New("invalid entry")

type Order uint8

const (
	Asc Order = iota
	Desc
)

func (_ Order) IsSelectOption() {}

func (_ Order) IsIterOption() {}

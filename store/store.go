// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/reusee/june/key"
	"github.com/reusee/sb"
)

type Key = key.Key

var (
	is = errors.Is
	pt = fmt.Printf
)

type ID string

type Store interface {
	ID(context.Context) (ID, error)
	Name() string

	Write(
		context.Context,
		key.Namespace,
		sb.Stream,
		...WriteOption,
	) (
		WriteResult,
		error,
	)

	Read(
		context.Context,
		Key,
		func(sb.Stream) error,
	) error

	Exists(
		context.Context,
		Key,
	) (bool, error)

	IterKeys(
		context.Context,
		key.Namespace,
		func(Key) error,
	) error

	IterAllKeys(
		context.Context,
		func(Key) error,
	) error

	Delete(
		context.Context,
		[]Key,
	) error
}

type WriteResult struct {
	Key          Key
	Written      bool
	BytesWritten int64
}

var Break = errors.New("break")

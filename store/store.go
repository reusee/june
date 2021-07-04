// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package store

import (
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
	ID() (ID, error)
	Name() string
	Close() error

	Sync() error

	Write(
		key.Namespace,
		sb.Stream,
		...WriteOption,
	) (
		WriteResult,
		error,
	)

	Read(
		Key,
		func(sb.Stream) error,
	) error

	Exists(
		Key,
	) (bool, error)

	IterKeys(
		key.Namespace,
		func(Key) error,
	) error

	IterAllKeys(
		fn func(Key) error,
	) error

	Delete(
		[]Key,
	) error
}

type WriteResult struct {
	Key          Key
	Written      bool
	BytesWritten int64
}

var Break = errors.New("break")

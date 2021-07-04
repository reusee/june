// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package store

import (
	"context"
	"errors"
)

var (
	ErrKeyNotFound   = errors.New("key not found")
	ErrKeyNotMatch   = errors.New("key not match")
	ErrRead          = errors.New("read error")
	ErrIgnore        = errors.New("ignore")
	ErrReadDisabled  = errors.New("read disabled")
	ErrWriteDisabled = errors.New("write disabled")
	ErrClosed        = context.Canceled
)

// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package store

import (
	"github.com/reusee/sb"
)

type Cache interface {
	CacheGet(Key, func(sb.Stream) error) error

	// CachePut must save tokens as encoded form
	CachePut(Key, sb.Tokens) error
}

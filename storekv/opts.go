// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storekv

import "github.com/reusee/june/store"

type withCodec struct {
	Codec Codec
}

func (withCodec) IsNewOption() {}

func WithCodec(codec Codec) withCodec {
	return withCodec{
		Codec: codec,
	}
}

type withCache struct {
	Cache Cache
}

func (withCache) IsNewOption() {}

func WithCache(cache Cache) withCache {
	return withCache{
		Cache: cache,
	}
}

type WithoutRead struct{}

func (WithoutRead) IsNewOption() {}

type WithoutWrite struct{}

func (WithoutWrite) IsNewOption() {}

type WithOffload func(
	key Key,
	length int,
) store.Store

func (WithOffload) IsNewOption() {}

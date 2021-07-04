// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storekv

type withCodec struct {
	Codec Codec
}

func (_ withCodec) IsNewOption() {}

func WithCodec(codec Codec) withCodec {
	return withCodec{
		Codec: codec,
	}
}

type withCache struct {
	Cache Cache
}

func (_ withCache) IsNewOption() {}

func WithCache(cache Cache) withCache {
	return withCache{
		Cache: cache,
	}
}

type WithoutRead struct{}

func (_ WithoutRead) IsNewOption() {}

type WithoutWrite struct{}

func (_ WithoutWrite) IsNewOption() {}

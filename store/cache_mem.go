// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package store

import (
	"bytes"

	lru "github.com/hashicorp/golang-lru"
	"github.com/reusee/e4"
	"github.com/reusee/june/key"
	"github.com/reusee/sb"
)

type MemCache struct {
	cache        *lru.Cache
	newHashState key.NewHashState
	maxSize      int
}

type NewMemCache func(
	maxKeys int,
	maxSize int,
) (*MemCache, error)

func (_ Def) NewMemCache(
	newHashState key.NewHashState,
) NewMemCache {
	return func(
		maxKeys int,
		maxSize int,
	) (_ *MemCache, err error) {
		defer he(&err)
		l, err := lru.New(maxKeys)
		ce(err)
		return &MemCache{
			cache:        l,
			newHashState: newHashState,
			maxSize:      maxSize,
		}, nil
	}
}

var _ Cache = new(MemCache)

func (m *MemCache) CacheGet(key Key, fn func(sb.Stream) error) (err error) {
	defer he(&err)
	v, ok := m.cache.Get(key)
	if !ok {
		return we(e4.With(key))(ErrKeyNotFound)
	}
	err = fn(sb.Decode(bytes.NewReader(v.([]byte))))
	ce(err)
	return nil
}

func (m *MemCache) CachePut(
	key Key,
	tokens sb.Tokens,
	options ...CachePutOption,
) error {
	if key.Valid() {
		if m.cache.Contains(key) {
			return nil
		}
	}

	var encodedLen *int
	for _, option := range options {
		switch option := option.(type) {
		case EncodedLen:
			l := int(option)
			encodedLen = &l
		}
	}

	if m.maxSize > 0 {
		var size int
		if encodedLen != nil {
			size = *encodedLen
		} else {
			ce(sb.Copy(
				tokens.Iter(),
				sb.EncodedLen(&size, nil),
			))
		}
		if size > m.maxSize {
			// ignore
			return nil
		}
	}

	var hash []byte
	buf := new(bytes.Buffer)
	if err := sb.Copy(
		tokens.Iter(),
		sb.Hash(m.newHashState, &hash, nil),
		sb.Encode(buf),
	); err != nil {
		return err
	}
	if !bytes.Equal(hash, key.Hash[:]) {
		panic("bad cache key")
	}
	m.cache.ContainsOrAdd(key, buf.Bytes())
	return nil
}

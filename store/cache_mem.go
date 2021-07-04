// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package store

import (
	"bytes"
	"fmt"

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
		return we(ErrKeyNotFound, e4.With(key))
	}
	err = fn(sb.Decode(bytes.NewReader(v.([]byte))))
	ce(err)
	return nil
}

func (m *MemCache) CachePut(key Key, tokens sb.Tokens) error {
	if key.Valid() {
		if m.cache.Contains(key) {
			return nil
		}
	}

	if m.maxSize > 0 {
		size := 0
		for _, token := range tokens {
			size++
			if token.Value != nil {
				switch v := token.Value.(type) {
				case bool, int8, uint8:
					size++
				case int16, uint16:
					size += 2
				case int32, uint32, float32:
					size += 4
				case int, uint, int64, uint64, float64:
					size += 8
				case string:
					size += 16 + len(v)
				case []byte:
					size += 24 + len(v)
				default:
					panic(fmt.Errorf("%T not handled", v))
				}
			}
			if size > m.maxSize {
				// ignore
				return nil
			}
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

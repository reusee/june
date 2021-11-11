// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storenssharded

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/reusee/e4"
	"github.com/reusee/june/key"
	"github.com/reusee/june/store"
	"github.com/reusee/sb"
)

type Store struct {
	name         string
	defaultStore store.Store
	shards       map[key.Namespace]store.Store
	set          map[store.Store]bool
	id           StoreID
}

type New func(
	shards map[key.Namespace]store.Store,
	def store.Store,
) (*Store, error)

func (_ Def) New() New {
	return func(
		shards map[key.Namespace]store.Store,
		def store.Store,
	) (_ *Store, err error) {
		defer he(&err)

		set := make(map[store.Store]bool)
		for _, s := range shards {
			set[s] = true
		}
		set[def] = true

		var ids []StoreID
		for s := range set {
			id, err := s.ID()
			ce(err)
			ids = append(ids, id)
		}
		sort.Slice(ids, func(i, j int) bool {
			return ids[i] < ids[j]
		})
		var b strings.Builder
		b.WriteString("(namespace-sharded(")
		for i, id := range ids {
			if i > 0 {
				b.WriteString(",")
			}
			b.WriteString(string(id))
		}
		b.WriteString("))")
		id := StoreID(b.String())

		storeNames := []string{
			def.Name(),
		}
		for _, store := range shards {
			storeNames = append(storeNames, store.Name())
		}

		return &Store{
			name: fmt.Sprintf("nssharded%d(%s)",
				atomic.AddInt64(&serial, 1),
				strings.Join(storeNames, ", "),
			),
			id:           id,
			shards:       shards,
			defaultStore: def,
			set:          set,
		}, nil
	}
}

var serial int64

var _ store.Store = new(Store)

func (s *Store) Name() string {
	return s.name
}

func (s *Store) ID() (StoreID, error) {
	return s.id, nil
}

func (s *Store) Exists(key Key) (bool, error) {
	if store, ok := s.shards[key.Namespace]; ok {
		return store.Exists(key)
	}
	return s.defaultStore.Exists(key)
}

func (s *Store) IterAllKeys(fn func(Key) error) (err error) {
	defer he(&err)
	var nsSet sync.Map
	var stop int64
	if err := s.defaultStore.IterAllKeys(func(key Key) (err error) {
		defer he(&err)
		if _, ok := nsSet.Load(key.Namespace); !ok {
			nsSet.Store(key.Namespace, true)
		}
		ce(
			fn(key),
			e4.WrapFunc(func(err error) error {
				if is(err, store.Break) {
					atomic.CompareAndSwapInt64(&stop, 0, 1)
				}
				return err
			}),
		)
		return nil
	}); err != nil {
		return err
	}
	if atomic.LoadInt64(&stop) > 0 {
		return nil
	}
	for ns, shard := range s.shards {
		ns := ns
		shard := shard
		if _, ok := nsSet.Load(ns); ok {
			continue
		}
		if err := shard.IterKeys(ns, func(key Key) error {
			if err := fn(key); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) IterKeys(ns key.Namespace, fn func(Key) error) error {
	if store, ok := s.shards[ns]; ok {
		return store.IterKeys(ns, fn)
	}
	return s.defaultStore.IterKeys(ns, fn)
}

func (s *Store) Read(key Key, fn func(sb.Stream) error) error {
	if store, ok := s.shards[key.Namespace]; ok {
		return store.Read(key, fn)
	}
	return s.defaultStore.Read(key, fn)
}

func (s *Store) Write(ns key.Namespace, stream sb.Stream, options ...WriteOption) (WriteResult, error) {
	if store, ok := s.shards[ns]; ok {
		return store.Write(ns, stream, options...)
	}
	return s.defaultStore.Write(ns, stream, options...)
}

func (s *Store) Delete(keys []Key) error {
	byNS := make(map[key.Namespace][]Key)
	for _, key := range keys {
		byNS[key.Namespace] = append(byNS[key.Namespace], key)
	}
	for ns, keys := range byNS {
		if store, ok := s.shards[ns]; ok {
			if err := store.Delete(keys); err != nil {
				return err
			}
		}
		if err := s.defaultStore.Delete(keys); err != nil {
			return err
		}
	}
	return nil
}

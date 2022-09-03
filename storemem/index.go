// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storemem

import (
	"fmt"
	"io"
	"sync/atomic"

	"github.com/reusee/e4"
	"github.com/reusee/june/index"
	"github.com/reusee/pp"
	"github.com/reusee/sb"
)

var _ index.IndexManager = new(Store)

func (s *Store) IndexFor(id StoreID) (index.Index, error) {
	return Index{
		name: fmt.Sprintf("mem-index%d(%s, %v)",
			atomic.AddInt64(&indexSerial, 1),
			s.Name(),
			id,
		),
		store: s,
		id:    id,
	}, nil
}

var indexSerial int64

type Index struct {
	name  string
	store *Store
	id    StoreID
}

func (i Index) Name() string {
	return i.name
}

func (i Index) Save(entry IndexEntry, options ...index.SaveOption) (err error) {
	defer he(&err)

	if entry.Type == nil {
		return we.With(
			e4.NewInfo("entry type is nil: %+v", entry),
		)(index.ErrInvalidEntry)
	}
	if entry.Key == nil && entry.Path == nil {
		return we.With(
			e4.NewInfo("both entry key and path is nil: %+v", entry),
		)(index.ErrInvalidEntry)
	}

	var tapEntry []IndexTapEntry
	for _, option := range options {
		switch option := option.(type) {
		case IndexTapEntry:
			tapEntry = append(tapEntry, option)
		default:
			panic(fmt.Errorf("unknown option: %T", option))
		}
	}

	tokens, err := sb.TokensFromStream(
		sb.Marshal(StoreIndex{
			ID:    i.id,
			Value: sb.Marshal(entry),
		}),
	)
	ce(err)
	i.store.Lock()
	i.store.index.ReplaceOrInsert(tokens)
	i.store.Unlock()

	if entry.Key != nil {
		preTokens, err := sb.TokensFromStream(
			sb.Marshal(StoreIndex{
				ID: i.id,
				Value: sb.Marshal(index.PreEntry{
					Key:   *entry.Key,
					Type:  entry.Type,
					Tuple: entry.Tuple,
				}),
			}),
		)
		ce(err)
		i.store.Lock()
		i.store.index.ReplaceOrInsert(preTokens)
		i.store.Unlock()
	}

	for _, tap := range tapEntry {
		tap(entry)
	}

	return nil
}

func (i Index) Delete(entry IndexEntry) (err error) {
	defer he(&err)
	tokens, err := sb.TokensFromStream(
		sb.Marshal(index.StoreIndex{
			ID:    i.id,
			Value: sb.Marshal(entry),
		}),
	)
	ce(err)
	i.store.Lock()
	i.store.index.Delete(tokens)
	i.store.Unlock()

	if entry.Key != nil {
		preTokens, err := sb.TokensFromStream(
			sb.Marshal(index.StoreIndex{
				ID: i.id,
				Value: sb.Marshal(index.PreEntry{
					Key:   *entry.Key,
					Type:  entry.Type,
					Tuple: entry.Tuple,
				}),
			}),
		)
		ce(err)
		i.store.Lock()
		i.store.index.Delete(preTokens)
		i.store.Unlock()
	}

	return nil
}

type indexIter struct {
	store   *Store
	lower   sb.Tokens
	upper   sb.Tokens
	current sb.Tokens
	order   Order
}

func (i Index) Iter(
	lower *sb.Tokens,
	upper *sb.Tokens,
	order Order,
) (
	_ pp.Src,
	_ io.Closer,
	err error,
) {
	defer he(&err)

	iter := &indexIter{
		store: i.store,
		order: order,
	}

	var lowerTokens sb.Tokens
	if lower == nil {
		lowerTokens = sb.MustTokensFromStream(
			sb.Marshal(StoreIndex{
				ID:    i.id,
				Value: sb.Marshal(sb.Min),
			}),
		)
	} else {
		lowerTokens, err = sb.TokensFromStream(
			sb.Marshal(StoreIndex{
				ID:    i.id,
				Value: lower.Iter(),
			}),
		)
		ce(err)
	}
	iter.lower = lowerTokens

	var upperTokens sb.Tokens
	if upper == nil {
		upperTokens = sb.MustTokensFromStream(
			sb.Marshal(StoreIndex{
				ID:    i.id,
				Value: sb.Marshal(sb.Max),
			}),
		)
	} else {
		upperTokens, err = sb.TokensFromStream(
			sb.Marshal(StoreIndex{
				ID:    i.id,
				Value: upper.Iter(),
			}),
		)
		ce(err)
	}
	iter.upper = upperTokens

	if order == Asc {
		iter.current = iter.lower
	} else {
		iter.current = iter.upper
	}

	return iter.Iter, iter, nil
}

func (m *indexIter) Close() error {
	return nil
}

func extractIndex(s sb.Stream) (sb.Stream, error) {
	var tokens sb.Tokens
	tuple := sb.Tuple{
		nil,
		sb.CollectValueTokens(&tokens),
	}
	if err := sb.Copy(
		s,
		sb.Unmarshal(&tuple),
	); err != nil {
		return nil, err
	}
	return tokens.Iter(), nil
}

func (m *indexIter) Iter() (_ any, _ pp.Src, err error) {
	defer he(&err)

	if m.order == Asc {
		n := 0
		var s sb.Stream
		m.store.RLock()
		defer m.store.RUnlock()
		m.store.index.AscendGreaterOrEqual(m.current, func(tokens sb.Tokens) bool {
			if sb.MustCompare(
				tokens.Iter(),
				m.upper.Iter(),
			) >= 0 {
				return false
			}
			if n == 0 {
				var err error
				s, err = extractIndex(tokens.Iter())
				ce(err)
				n++
				return true
			} else if n == 1 {
				m.current = tokens
				n++
				return false
			}
			return false
		})
		if n == 0 {
			return nil, nil, nil
		} else if n == 1 {
			return s, nil, nil
		} else {
			return s, m.Iter, nil
		}

	} else {
		n := 0
		var s sb.Stream
		m.store.RLock()
		defer m.store.RUnlock()
		m.store.index.DescendLessOrEqual(m.current, func(tokens sb.Tokens) bool {
			if sb.MustCompare(
				tokens.Iter(),
				m.upper.Iter(),
			) == 0 {
				// exclude upper
				return true
			}
			if sb.MustCompare(
				tokens.Iter(),
				m.lower.Iter(),
			) < 0 {
				return false
			}
			if n == 0 {
				var err error
				s, err = extractIndex(tokens.Iter())
				ce(err)
				n++
				return true
			} else if n == 1 {
				m.current = tokens
				n++
				return false
			}
			return false
		})
		if n == 0 {
			return nil, nil, nil
		} else if n == 1 {
			return s, nil, nil
		} else {
			return s, m.Iter, nil
		}

	}

}

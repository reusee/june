// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package entity

import (
	"context"
	"fmt"
	"reflect"

	"github.com/reusee/dscope"
	"github.com/reusee/june/index"
	"github.com/reusee/june/key"
	"github.com/reusee/june/store"
	"github.com/reusee/sb"
)

type Summary struct {
	Key         Key
	SlotHash    *Hash
	Indexes     *[]IndexEntry // IndexEntry without Key in Tuple
	Subs        []*Summary
	ReferedKeys []Key // all in NSEntity

	// update following methods after adding fields
	// 如果增加了字段，也应修改如下方法
	// clean
	// checkRef
}

var _ sb.HasDeprecatedFields = Summary{}

func (s Summary) SBDeprecatedFields() []string {
	return []string{"RawIndexes"}
}

var NSSummary = key.Namespace{'c', '-', 's', 'u', 'm', 'a', 'r', 'y'}

var _ sb.SBMarshaler = Summary{}

func (s Summary) MarshalSB(ctx sb.Ctx, cont sb.Proc) sb.Proc {
	ctx.SkipEmptyStructFields = true
	return sb.MarshalStruct(ctx, reflect.ValueOf(s), cont)
}

func (s Summary) Valid() bool {
	return s.Key.Valid()
}

func (s *Summary) clean() (ok bool) {

	// SlotHash
	if s.SlotHash != nil {
		ok = true
	}

	// Indexes
	if s.Indexes != nil && len(*s.Indexes) > 0 {
		ok = true
	}

	// Subs
	newSubs := s.Subs[:0]
	for _, sub := range s.Subs {
		if ok := sub.clean(); ok {
			newSubs = append(newSubs, sub)
		}
	}
	s.Subs = newSubs
	if len(s.Subs) > 0 {
		ok = true
	}

	// ReferedKeys
	if len(s.ReferedKeys) > 0 {
		ok = true
	}

	return
}

type OnSummaryIndexAdd func(
	ctx context.Context,
	summary *Summary,
	summaryKey Key,
) (
	entries []IndexEntry,
	err error,
)

var _ dscope.CustomReducer = OnSummaryIndexAdd(nil)

func (o OnSummaryIndexAdd) Reduce(_ dscope.Scope, vs []reflect.Value) reflect.Value {
	var fns []OnSummaryIndexAdd
	for _, v := range vs {
		fns = append(fns, v.Interface().(OnSummaryIndexAdd))
	}
	fn := OnSummaryIndexAdd(func(
		ctx context.Context,
		summary *Summary,
		summaryKey Key,
	) (
		entries []IndexEntry,
		err error,
	) {
		for _, f := range fns {
			if es, err := f(ctx, summary, summaryKey); err != nil {
				return nil, err
			} else {
				entries = append(entries, es...)
			}
		}
		return
	})
	return reflect.ValueOf(fn)
}

type OnSummaryIndexDelete func(
	ctx context.Context,
	summary *Summary,
	summaryKey Key,
) (
	entries []IndexEntry,
	err error,
)

var _ dscope.CustomReducer = OnSummaryIndexDelete(nil)

func (o OnSummaryIndexDelete) Reduce(_ dscope.Scope, vs []reflect.Value) reflect.Value {
	var fns []OnSummaryIndexDelete
	for _, v := range vs {
		fns = append(fns, v.Interface().(OnSummaryIndexDelete))
	}
	fn := OnSummaryIndexDelete(func(
		ctx context.Context,
		summary *Summary,
		summaryKey Key,
	) (
		entris []IndexEntry,
		err error,
	) {
		for _, f := range fns {
			if es, err := f(ctx, summary, summaryKey); err != nil {
				return nil, err
			} else {
				entris = append(entris, es...)
			}
		}
		return
	})
	return reflect.ValueOf(fn)
}

type idxEmbeddedBy struct {
	EmbeddedKey Key
}

var IdxEmbeddedBy = idxEmbeddedBy{}

func init() {
	index.Register(IdxEmbeddedBy)
}

func (Def) SummaryIndexFuncs(
	sel index.SelectIndex,
) (
	add OnSummaryIndexAdd,
	del OnSummaryIndexDelete,
) {

	add = func(
		_ context.Context,
		summary *Summary,
		_ Key,
	) (
		entries []IndexEntry,
		err error,
	) {
		defer he(&err)
		ce(summary.iterAll(func(path []Key, subSummary Summary) (err error) {
			defer he(&err)
			if subSummary.Indexes == nil {
				return
			}
			for _, idx := range *subSummary.Indexes {
				key := path[len(path)-1]
				entry := idx.Clone()
				entry.Key = &key
				entries = append(entries, entry)
				if len(path) > 1 {
					entries = append(entries,
						NewEntry(IdxEmbeddedBy, key, summary.Key),
					)
				}
			}
			return nil
		}))
		return
	}

	del = func(
		ctx context.Context,
		summary *Summary,
		_ Key,
	) (
		entries []IndexEntry,
		err error,
	) {
		defer he(&err)
		ce(summary.iterAll(func(path []Key, subSummary Summary) (err error) {
			defer he(&err)
			if subSummary.Indexes == nil {
				return
			}
			for _, idx := range *subSummary.Indexes {
				key := path[len(path)-1]
				if len(path) > 1 {
					// embedded
					entries = append(entries,
						NewEntry(IdxEmbeddedBy, key, summary.Key),
					)
					//  check ref
					var n int
					ce(sel(
						ctx,
						MatchEntry(IdxEmbeddedBy, key),
						Count(&n),
					))
					if n == 1 {
						// last ref
						entry := idx.Clone()
						entry.Key = &key
						entries = append(entries, entry)
					}
				} else {
					// not embedded
					entry := idx.Clone()
					entry.Key = &key
					entries = append(entries, entry)
				}
			}
			return nil
		}))
		return
	}

	return
}

type SaveSummaryOption interface {
	IsSaveSummaryOption()
}

func (WithIndexSaveOptions) IsSaveSummaryOption() {}

// SaveSummary
type SaveSummary func(
	ctx context.Context,
	summary *Summary,
	isLatest bool,
	options ...SaveSummaryOption,
) error

func (Def) SaveSummary(
	store store.Store,
	sel index.SelectIndex,
	fetch Fetch,
	index Index,
	locks *_EntityLocks,
	onAdd OnSummaryIndexAdd,
	onDel OnSummaryIndexDelete,
) SaveSummary {

	return func(
		ctx context.Context,
		s *Summary,
		isLatest bool,
		options ...SaveSummaryOption,
	) (
		err error,
	) {
		defer he(&err)

		// lock with entity key
		unlock := locks.Lock(s.Key)
		defer unlock()

		//TODO lock slot

		// options
		var tapKey []TapKey
		var indexSaveOptions []IndexSaveOption
		for _, option := range options {
			switch option := option.(type) {
			case TapKey:
				tapKey = append(tapKey, option)
			case WithIndexSaveOptions:
				indexSaveOptions = append(indexSaveOptions, option...)
			default:
				panic(fmt.Errorf("bad option: %T", option))
			}
		}

		// clean
		s.clean()

		// save
		proc := sb.MarshalValue(sb.Ctx{
			SkipEmptyStructFields: true,
		}, reflect.ValueOf(s), nil)
		res, err := store.Write(ctx, NSSummary, &proc)
		ce(err)
		summaryKey := res.Key

		for _, fn := range tapKey {
			fn(summaryKey)
		}

		// remove old summaries
		var oldSummaryKeys []Key
		var oldSummaries []Summary
		if isLatest {
			ce(sel(
				ctx,
				MatchEntry(IdxSummaryKey, s.Key),
				TapKey(func(k Key) {
					if k == summaryKey {
						return
					}
					var oldSummary Summary
					err := fetch(ctx, k, &oldSummary)
					if is(err, ErrKeyNotFound) {
						return
					}
					oldSummaryKeys = append(oldSummaryKeys, k)
					oldSummaries = append(oldSummaries, oldSummary)
				}),
			))
			for _, oldKey := range oldSummaryKeys {
				ce(store.Delete(ctx, []Key{oldKey}))
			}
		}

		//TODO also remove old slot entities

		// update indexes
		if len(oldSummaryKeys) > 0 {
			deletingIndexes := make(map[Hash]IndexEntry)
			for i, oldKey := range oldSummaryKeys {
				entries, err := onDel(ctx, &oldSummaries[i], oldKey)
				ce(err)
				for _, entry := range entries {
					h, err := key.HashValue(entry)
					ce(err)
					deletingIndexes[h] = entry
				}
			}
			entries, err := onAdd(ctx, s, summaryKey)
			ce(err)
			for _, entry := range entries {
				h, err := key.HashValue(entry)
				ce(err)
				if _, ok := deletingIndexes[h]; ok {
					delete(deletingIndexes, h)
					continue
				}
				ce(index.Save(ctx, entry, indexSaveOptions...))
			}
			for _, entry := range deletingIndexes {
				ce(index.Delete(ctx, entry))
			}

		} else {
			entries, err := onAdd(ctx, s, summaryKey)
			ce(err)
			for _, entry := range entries {
				ce(index.Save(ctx, entry, indexSaveOptions...))
			}
		}

		return nil
	}
}

func (s Summary) iterAll(fn func([]Key, Summary) error) error {
	return s._iterAll(nil, fn)
}

func (s Summary) _iterAll(path []Key, fn func([]Key, Summary) error) (err error) {
	defer he(&err)
	path = append(path, s.Key)
	ce(fn(path, s))
	for _, sub := range s.Subs {
		ce(sub._iterAll(path, fn))
	}
	return nil
}

func (s *Summary) addIndex(entry IndexEntry) error {
	if s.Indexes == nil {
		s.Indexes = &[]IndexEntry{}
	}
	*s.Indexes = append(*s.Indexes, entry)
	return nil
}

type DeleteSummary func(
	ctx context.Context,
	summary *Summary,
	summaryKey Key,
) (
	err error,
)

func (Def) DeleteSummary(
	store store.Store,
	index Index,
	locks *_EntityLocks,
	onDel OnSummaryIndexDelete,
) DeleteSummary {
	return func(
		ctx context.Context,
		summary *Summary,
		summaryKey Key,
	) (
		err error,
	) {
		defer he(&err)

		unlock := locks.Lock(summary.Key)
		defer unlock()

		//TODO lock slot

		// indexes
		tuples, err := onDel(ctx, summary, summaryKey)
		ce(err)
		for _, tuple := range tuples {
			ce(index.Delete(ctx, tuple))
		}

		// delete
		ce(store.Delete(ctx, []Key{summaryKey}))

		return
	}
}

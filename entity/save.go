// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package entity

import (
	"bytes"
	"fmt"
	"reflect"
	"sort"

	"github.com/reusee/dscope"
	"github.com/reusee/june/key"
	"github.com/reusee/june/naming"
	"github.com/reusee/june/store"
	"github.com/reusee/sb"
)

// save entity
type Save func(
	ns key.Namespace,
	value any,
	options ...SaveOption,
) (
	summary *Summary,
	err error,
)

type SaveOption interface {
	IsSaveOption()
}

type SaveSummaryOptions []SaveSummaryOption

func (_ SaveSummaryOptions) IsSaveOption() {}

type SaveValue any

func (_ Def) Save(
	store store.Store,
	scope dscope.Scope,
	newHashState key.NewHashState,
	saveSummary SaveSummary,
) (
	save Save,
) {

	keyType := reflect.TypeOf((*Key)(nil)).Elem()

	save = func(
		ns key.Namespace,
		value any,
		options ...SaveOption,
	) (summary *Summary, err error) {
		defer he(&err)

		/*
		  实现需保证：
		  所有事实保存于store中
		  需要索引的数据，都保存于summary中
		  所有索引基于summary的内容，且可以随时删除并重建
		  连续插入相同value，store不产生变更
		*/

		// options
		var tapSummary TapSummary
		var tapWriteResult TapWriteResult
		var saveSummaryOptions []SaveSummaryOption
		for _, option := range options {
			switch option := option.(type) {
			case TapSummary:
				tapSummary = option
			case TapWriteResult:
				tapWriteResult = option
			case SaveSummaryOptions:
				saveSummaryOptions = append(saveSummaryOptions, option...)
			default:
				panic(fmt.Errorf("unknown option: %T", option))
			}
		}

		typeName := naming.Type(reflect.TypeOf(value))

		// marshal
		var stack []*Summary
		var cur *Summary
		referedKeys := make(map[Key]struct{})
		fn := func(ctx sb.Ctx, value reflect.Value, cont sb.Proc) sb.Proc {

			// referenced key
			if value.Type() == keyType {
				referedKeys[value.Interface().(Key)] = struct{}{}
			}

			// summary
			cur = new(Summary)

			// index
			if value.Type().Implements(hasIndexType) {
				indexes, version, err := value.Interface().(HasIndex).EntityIndexes()
				if err != nil {
					return func() (*sb.Token, sb.Proc, error) {
						return nil, nil, err
					}
				}
				ce(cur.addIndex(IdxVersion(version)))
				for _, index := range indexes {
					if err := cur.addIndex(index); err != nil {
						return func() (*sb.Token, sb.Proc, error) {
							return nil, nil, err
						}
					}
				}
			}

			return sb.MarshalValue(ctx, value, cont)
		}
		proc := fn(sb.Ctx{
			Marshal: fn,
		}, reflect.ValueOf(value), nil)

		// hash
		var rootSummary *Summary
		var sum []byte
		proc = sb.Tee(
			proc,
			sb.HashFunc(
				newHashState,
				&sum,
				func(hash []byte, token *sb.Token) error {

					// must maintain stack here
					if len(hash) > 0 {
						copy(
							stack[len(stack)-1].Key.Hash[:],
							hash,
						)
						if len(stack) == 1 {
							rootSummary = stack[0]
						}
						stack = stack[:len(stack)-1]

					} else {
						if len(stack) > 0 {
							parent := stack[len(stack)-1]
							parent.Subs = append(parent.Subs, cur)
						}
						stack = append(stack, cur)
					}

					return nil
				},
				nil,
			),
		)

		// write
		var res WriteResult
		res, err = store.Write(ns, proc)
		if tapWriteResult != nil {
			tapWriteResult(res)
		}
		ce(err)
		if !bytes.Equal(sum, res.Key.Hash[:]) {
			return nil, fmt.Errorf("hash not match")
		}

		rootSummary.Key = res.Key

		var keys []Key
		for key := range referedKeys {
			keys = append(keys, key)
		}
		sort.Slice(keys, func(i, j int) bool {
			return bytes.Compare(keys[i].Hash[:], keys[j].Hash[:]) < 0
		})
		rootSummary.ReferedKeys = keys

		// type name index
		ce(rootSummary.addIndex(IdxType(typeName)))

		// save summary
		ce(saveSummary(rootSummary, true, saveSummaryOptions...))

		if tapSummary != nil {
			tapSummary(rootSummary)
		}

		return rootSummary, nil
	}

	return
}

type SaveEntity func(value any, options ...SaveOption) (summary *Summary, err error)

func (_ Def) SaveEntity(
	save Save,
) SaveEntity {
	return func(value any, options ...SaveOption) (*Summary, error) {
		return save(NSEntity, value, options...)
	}
}

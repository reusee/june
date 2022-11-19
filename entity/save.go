// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package entity

import (
	"bytes"
	"context"
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
	ctx context.Context,
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

func (SaveSummaryOptions) IsSaveOption() {}

type SaveValue any

func (Def) Save(
	store store.Store,
	scope dscope.Scope,
	newHashState key.NewHashState,
	saveSummary SaveSummary,
) (
	save Save,
) {

	keyType := reflect.TypeOf((*Key)(nil)).Elem()

	save = func(
		ctx context.Context,
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
			valueType := value.Type()

			// referenced key
			if valueType == keyType {
				referedKeys[value.Interface().(Key)] = struct{}{}
			}

			// summary
			cur = new(Summary)

			// index
			if valueType.Implements(hasIndexType) {
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

			// slot
			if valueType.Implements(hasSlotKeysType) {
				keys, err := value.Interface().(HasSlotKeys).SlotKeys()
				if err != nil {
					return func() (*sb.Token, sb.Proc, error) {
						return nil, nil, err
					}
				}
				var hash []byte
				if err := sb.Copy(
					sb.Marshal(keys),
					sb.Hash(newHashState, &hash, nil),
				); err != nil {
					return func() (*sb.Token, sb.Proc, error) {
						return nil, nil, err
					}
				}
				var h Hash
				copy(h[:], hash)
				cur.SlotHash = &h
			}

			return sb.MarshalValue(ctx, value, cont)
		}
		proc := fn(sb.Ctx{
			Marshal: fn,
		}, reflect.ValueOf(value), nil)
		stream := &proc

		// hash
		var rootSummary *Summary
		var sum []byte
		stream = sb.Tee(
			stream,
			sb.HashFunc(
				newHashState,
				&sum,
				func(hash []byte, _ *sb.Token) error {

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
		res, err = store.Write(ctx, ns, stream)
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
		ce(saveSummary(ctx, rootSummary, true, saveSummaryOptions...))

		if tapSummary != nil {
			tapSummary(rootSummary)
		}

		return rootSummary, nil
	}

	return
}

type SaveEntity func(ctx context.Context, value any, options ...SaveOption) (summary *Summary, err error)

func (Def) SaveEntity(
	save Save,
) SaveEntity {
	return func(ctx context.Context, value any, options ...SaveOption) (*Summary, error) {
		return save(ctx, NSEntity, value, options...)
	}
}

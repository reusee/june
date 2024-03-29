// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package entity

import (
	"context"
	"fmt"

	"github.com/reusee/dscope"
	"github.com/reusee/june/codec"
	"github.com/reusee/june/storekv"
	"github.com/reusee/june/storemem"
	"github.com/reusee/june/sys"
	"github.com/reusee/pr2"
	"github.com/reusee/sb"
)

type IndexGC func(
	ctx context.Context,
	options ...IndexGCOption,
) error

type IndexGCOption interface {
	IsIndexGCOption()
}

type blackholeCodec struct{}

var _ codec.Codec = blackholeCodec{}

func (blackholeCodec) Encode(sink sb.Sink, options ...codec.Option) sb.Sink {
	return sb.Discard
}

func (blackholeCodec) Decode(str sb.Stream, options ...codec.Option) sb.Stream {
	panic("should not be called")
}

func (blackholeCodec) ID() string {
	return "blackhole"
}

func (Def) IndexGC(
	store Store,
	newMem storemem.New,
	scope dscope.Scope,
	newKV storekv.New,
	fetch Fetch,
	index Index,
	parallel sys.Parallel,
) IndexGC {
	return func(
		ctx context.Context,
		options ...IndexGCOption,
	) (err error) {
		defer he(&err)

		var tapDeleted []TapDeleteIndex
		for _, option := range options {
			switch option := option.(type) {
			case TapDeleteIndex:
				tapDeleted = append(tapDeleted, option)
			default:
				panic(fmt.Errorf("unknown option: %T", option))
			}
		}

		// rebuild summary in mem store
		memStore := newMem(ctx)
		memScope := scope.Fork(func() (Store, IndexManager) {
			kv, err := newKV(ctx, memStore, "index-gc", storekv.WithCodec(
				blackholeCodec{},
			))
			ce(err)
			return kv, memStore
		})

		memScope.Call(func(
			memSaveSummary SaveSummary,
			memIndex Index,
		) {

			// save
			ce(store.IterKeys(NSSummary, func(key Key) (err error) {
				defer he(&err)
				var summary Summary
				ce(fetch(key, &summary))
				ce(memSaveSummary(ctx, &summary, true))
				return nil
			}))

			// iter
			src, closeSrc, err := index.Iter(
				nil,
				nil,
				Asc,
			)
			ce(err)
			defer closeSrc.Close()

			// iter
			memSrc, closeMemSrc, err := memIndex.Iter(
				nil,
				nil,
				Asc,
			)
			ce(err)
			defer closeMemSrc.Close()

			next := func() (tokens sb.Tokens) {
				v, err := src.Next()
				ce(err)
				if v == nil {
					return
				}
				stream := v.(sb.Stream)
				ce(sb.Copy(
					stream,
					sb.CollectTokens(&tokens),
				))
				return
			}

			memNext := func() (tokens sb.Tokens) {
				v, err := memSrc.Next()
				ce(err)
				if v == nil {
					return
				}
				stream := v.(sb.Stream)
				ce(sb.Copy(
					stream,
					sb.CollectTokens(&tokens),
				))
				return
			}

			wg := pr2.NewWaitGroup(ctx)
			defer wg.Cancel()
			put, wait := pr2.Consume(wg, int(parallel), func(_ int, v any) (err error) {
				defer he(&err)
				tokens := v.(sb.Tokens)
				if tokens[1].Kind != sb.KindString {
					// non-Entry
					return nil
				}
				var entry IndexEntry
				ce(sb.Copy(
					tokens.Iter(),
					sb.Unmarshal(&entry),
				))
				ce(index.Delete(entry))
				for _, tap := range tapDeleted {
					tap(entry)
				}
				return nil
			})

			tokens := next()
			memTokens := memNext()
			for tokens != nil && memTokens != nil {
				res, err := sb.Compare(tokens.Iter(), memTokens.Iter())
				ce(err)
				switch res {

				case 0:
					tokens = next()
					memTokens = memNext()

				case 1:
					memTokens = memNext()

				case -1:
					put(tokens)
					tokens = next()

				}
			}

			for tokens != nil {
				put(tokens)
				tokens = next()
			}

			ce(closeSrc.Close())
			ce(closeMemSrc.Close())
			ce(wait(true))

		})

		return
	}
}

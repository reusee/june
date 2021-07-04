// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package entity

import (
	"context"
	"fmt"
	"sync"

	"github.com/reusee/ling/v2/index"
	"github.com/reusee/ling/v2/sys"
	"github.com/reusee/pr"
	"github.com/reusee/sb"
)

type GC func(
	roots []Key,
	options ...GCOption,
) error

type DeadObject struct {
	Key     Key
	Summary *Summary
}

type GCOption interface {
	IsGCOption()
}

func (_ Def) GC(
	store Store,
	selIndex index.SelectIndex,
	index Index,
	rootCtx context.Context,
	parallel sys.Parallel,
	deleteSummary DeleteSummary,
) GC {

	return func(
		roots []Key,
		options ...GCOption,
	) (err error) {
		defer he(&err)

		if len(roots) == 0 {
			return we(fmt.Errorf("no root"))
		}

		var tapMark TapMarkKey
		var tapReachable TapReachableObjects
		var tapIter TapIterKey
		var tapDead TapDeadObjects
		var tapSweep TapSweepDeadObject

		for _, option := range options {
			switch option := option.(type) {
			case TapMarkKey:
				tapMark = option
			case TapReachableObjects:
				tapReachable = option
			case TapIterKey:
				tapIter = option
			case TapDeadObjects:
				tapDead = option
			case TapSweepDeadObject:
				tapSweep = option
			default:
				panic(fmt.Errorf("unknown option: %T", option))
			}
		}

		// mark
		var reachable sync.Map // Key: struct{}

		ctx, cancel := context.WithCancel(rootCtx)
		defer cancel()
		var put pr.Put
		put, wait := pr.Consume(ctx, int(parallel), func(i int, v any) (err error) {
			defer he(&err)

			key := v.(Key)
			if _, ok := reachable.Load(key); ok {
				return nil
			}
			reachable.Store(key, struct{}{})
			if tapMark != nil {
				tapMark(key)
			}
			var keys []Key
			ce(Select(
				index,
				MatchEntry(IdxReferTo, key),
				Tap(func(_ Key, toKey Key) {
					keys = append(keys, toKey)
				}),
			))
			for _, key := range keys {
				put(key)
			}
			return nil
		})

		for _, key := range roots {
			put(key)
		}

		ce(wait(false))

		if tapReachable != nil {
			tapReachable(&reachable)
		}

		if len(roots) == 0 {
			panic("empty root")
		}

		// collect dead objects
		deadObjects := make([][]DeadObject, int(parallel))
		put, wait = pr.Consume(ctx, int(parallel), func(i int, v any) (err error) {
			defer he(&err)

			key := v.(Key)
			if tapIter != nil {
				tapIter(key)
			}
			if key.Namespace == NSSummary {
				var summary Summary
				ce(store.Read(key, func(stream sb.Stream) error {
					return sb.Copy(stream, sb.Unmarshal(&summary))
				}))
				if _, ok := reachable.Load(summary.Key); ok {
					return nil
				}
				deadObjects[i] = append(deadObjects[i], DeadObject{
					Key:     key,
					Summary: &summary,
				})
			} else {
				if _, ok := reachable.Load(key); ok {
					return nil
				}
				deadObjects[i] = append(deadObjects[i], DeadObject{
					Key: key,
				})
			}
			return nil
		})

		// must use index, to avoid empty index causing all objects to be deleted
		ce(selIndex(
			MatchEntry(IdxSummaryKey),
			Tap(func(key Key, summaryKey Key) {
				put(key)
				put(summaryKey)
			}),
		))
		ce(wait(true))

		var objs []DeadObject
		for _, os := range deadObjects {
			objs = append(objs, os...)
		}
		if tapDead != nil {
			tapDead(objs)
		}

		// delete
		batchKeys := make([][]Key, int(parallel))
		put, wait = pr.Consume(ctx, int(parallel), func(proc int, v any) (err error) {
			defer he(&err)

			obj := v.(DeadObject)
			if obj.Key.Namespace == NSSummary {
				ce(deleteSummary(obj.Summary, obj.Key))
			} else {
				if len(batchKeys[proc]) > 500 {
					ce(store.Delete(batchKeys[proc]))
					batchKeys[proc] = batchKeys[proc][:0]
				}
				batchKeys[proc] = append(batchKeys[proc], obj.Key)
			}
			if tapSweep != nil {
				tapSweep(obj)
			}
			return nil
		})
		for _, obj := range objs {
			put(obj)
		}
		ce(wait(true))
		for _, keys := range batchKeys {
			ce(store.Delete(keys))
		}

		return nil
	}
}

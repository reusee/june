// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package entity

import (
	"fmt"

	"github.com/reusee/june/index"
)

/*
TODO
增加垃圾清理选项
删除时，可以合并检查所引用的数据
如果只有单一引用，则可以删除，递归进行
与写操作的并发，可以统一做一个标记，然后由统一的过程来判断是否执行删除
就是如果有写标记，则不删除
*/

type Delete func(
	key Key,
) (
	err error,
)

type BeingRefered struct {
	Key Key
	By  map[Key]struct{}
}

var _ error = new(BeingRefered)

func (e *BeingRefered) Error() string {
	return fmt.Sprintf(
		"%s is refered by %+v",
		e.Key,
		e.By,
	)
}

func (_ Def) Delete(
	fetch Fetch,
	store Store,
	index Index,
	sel index.SelectIndex,
	deleteSummary DeleteSummary,
) Delete {

	//TODO entity lock

	return func(
		key Key,
	) (
		err error,
	) {
		defer he(&err)

		var entityKey Key
		summaryKeys := make(map[Key]struct{})
		if key.Namespace == NSSummary {
			summaryKeys[key] = struct{}{}
			ce(sel(
				MatchEntry(IdxSummaryOf, key),
				TapKey(func(key Key) {
					entityKey = key
				}),
			))
			ce(sel(
				MatchEntry(IdxSummaryKey, entityKey),
				TapKey(func(key Key) {
					summaryKeys[key] = struct{}{}
				}),
			))
		} else {
			entityKey = key
			ce(sel(
				MatchEntry(IdxSummaryKey, entityKey),
				TapKey(func(key Key) {
					summaryKeys[key] = struct{}{}
				}),
			))
		}

		// IdxReferedBy
		referring := make(map[Key]struct{})
		ce(Select(
			index,
			MatchEntry(IdxReferedBy, entityKey),
			Tap(func(_ Key, referringKey Key) error {
				if _, ok := summaryKeys[referringKey]; !ok {
					referring[referringKey] = struct{}{}
				}
				return nil
			}),
		))
		if len(referring) > 0 {
			return &BeingRefered{
				Key: key,
				By:  referring,
			}
		}

		// delete summary
		for summaryKey := range summaryKeys {
			var summary Summary
			ce(fetch(summaryKey, &summary))
			ce(deleteSummary(&summary, summaryKey))
		}

		// delete entity
		ce(store.Delete([]Key{
			entityKey,
		}))

		return
	}

}

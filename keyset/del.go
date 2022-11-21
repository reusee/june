// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package keyset

import (
	"sort"

	"github.com/reusee/june/entity"
)

type Delete func(
	set Set,
	keys ...Key,
) (
	newSet Set,
	err error,
)

func (_ Def) Delete(
	fetch entity.Fetch,
) Delete {
	return func(
		set Set,
		keys ...Key,
	) (
		newSet Set,
		err error,
	) {
		defer he(&err)

		for _, key := range keys {
			set, err = deleteKeyFromSet(fetch, key, set)
			ce(err)
		}

		newSet = set
		return
	}
}

func deleteKeyFromSet(
	fetch entity.Fetch,
	key Key,
	set Set,
) (
	newSet Set,
	err error,
) {
	defer he(&err)

	i := sort.Search(len(set), func(i int) bool {
		item := set[i]
		if item.Key != nil {
			return key.Compare(*item.Key) <= 0
		} else if item.Pack != nil {
			return key.Compare(item.Pack.Max) <= 0
		}
		panic("bad item")
	})

	if i >= len(set) {
		// not in set
		return set, nil
	}

	item := set[i]
	if item.Key != nil {
		if *item.Key == key {
			// delete
			newSet = make(Set, 0, len(set)-1)
			newSet = append(newSet, set[:i]...)
			newSet = append(newSet, set[i+1:]...)
			return newSet, nil

		} else {
			// not in set
			return set, nil
		}

	} else if item.Pack != nil {
		var s Set
		ce(fetch(item.Pack.Key, &s))
		replace, err := deleteKeyFromSet(
			fetch,
			key,
			s,
		)
		ce(err)
		newSet = make(Set, 0, len(set)+len(replace))
		newSet = append(newSet, set[:i]...)
		newSet = append(newSet, replace...)
		newSet = append(newSet, set[i+1:]...)
		return newSet, nil
	}

	panic("bad item")
}

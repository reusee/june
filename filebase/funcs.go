// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package filebase

import (
	"sort"
	"sync"

	"github.com/reusee/ling/v2/entity"
	"github.com/reusee/ling/v2/key"
	"github.com/reusee/pp"
)

type IterSubs func(subs Subs, cont pp.Src) pp.Src

type FindFileInSubs func(subs Subs, parts []string) (*File, error)

func (_ Def) Funcs(
	fetch entity.Fetch,
	newContentReader NewContentReader,
	newHashState key.NewHashState,
) (
	iterSubs IterSubs,
	findFileInSubs FindFileInSubs,
) {

	var subsCache sync.Map
	fetchSubs := func(key Key) (Subs, error) {
		if v, ok := subsCache.Load(key); ok {
			return v.(Subs), nil
		}
		var subs Subs
		if err := fetch(key, &subs); err != nil {
			return nil, err
		}
		subsCache.Store(key, subs)
		return subs, nil
	}

	iterSubs = func(subs Subs, cont pp.Src) pp.Src {
		var src pp.Src
		src = func() (_ any, _ pp.Src, err error) {
			defer he(&err)
			if len(subs) == 0 {
				return nil, cont, nil
			}
			sub := subs[0]
			subs = subs[1:]
			if sub.File != nil {
				return sub.File, src, nil
			} else if sub.Pack != nil {
				childSubs, err := fetchSubs(sub.Pack.Key)
				ce(err)
				return nil, iterSubs(childSubs, src), nil
			}
			panic("bad sub")
		}
		return src
	}

	findFileInSubs = func(subs Subs, parts []string) (_ *File, err error) {
		defer he(&err)

		if len(parts) == 0 || len(subs) == 0 {
			return nil, nil
		}
		name := parts[0]
		i := sort.Search(len(subs), func(i int) bool {
			sub := subs[i]
			if sub.File != nil {
				return sub.File.Name >= name
			} else if sub.Pack != nil {
				return sub.Pack.Max >= name
			}
			panic("bad sub")
		})
		if i >= len(subs) {
			return nil, nil
		}
		sub := subs[i]
		if sub.File != nil {
			if sub.File.Name != name {
				return nil, nil
			}
			if len(parts) == 1 {
				return sub.File, nil
			}
			return findFileInSubs(sub.File.Subs, parts[1:])
		} else if sub.Pack != nil {
			childSubs, err := fetchSubs(sub.Pack.Key)
			ce(err)
			return findFileInSubs(childSubs, parts)
		}
		panic("bad sub")
	}

	return
}

// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package keyset

import (
	"sort"

	"github.com/reusee/june/entity"
)

type Threshold int

func (_ Def) Threshold() Threshold {
	return 128
}

type Add func(
	set Set,
	keys ...Key,
) (
	Set,
	error,
)

func (_ Def) Add(
	fetch entity.Fetch,
	save entity.SaveEntity,
	threshold Threshold,
) Add {
	return func(
		set Set,
		keys ...Key,
	) (
		_ Set,
		err error,
	) {
		defer he(&err)

		for _, key := range keys {
			set, err = mergeKeyToSet(
				fetch,
				key,
				set,
			)
			ce(err)
		}

		set, err = pack(set, int(threshold), save)
		ce(err)

		return set, nil
	}
}

func mergeKeyToSet(
	fetchEntity entity.Fetch,
	key Key,
	set Set,
) (_ Set, err error) {
	defer he(&err)

	i := sort.Search(len(set), func(i int) bool {
		item := set[i]
		if item.Key != nil {
			return *item.Key == key
		} else if item.Pack != nil {
			return key.Compare(item.Pack.Key) <= 0
		}
		panic("invalid item") // NOCOVER
	})

	if i < len(set) {
		// found
		item := set[i]

		if (item.Key != nil && key.Compare(*item.Key) < 0) ||
			(item.Pack != nil && key.Compare(item.Pack.Min) < 0) {
			// insert
			newSet := make(Set, 0, len(set)+1)
			newSet = append(newSet, set[:i]...)
			newSet = append(newSet, SetItem{
				Key: &key,
			})
			newSet = append(newSet, set[i:]...)
			return newSet, nil

		} else if item.Key != nil && key == *item.Key {
			// existed
			return set, nil

		} else if item.Pack != nil &&
			key.Compare(item.Pack.Min) >= 0 &&
			key.Compare(item.Pack.Max) <= 0 {
			// merge to pack
			newSet := make(Set, 0, len(set))
			newSet = append(newSet, set[:i]...)
			replace, err := mergeKeyToPack(fetchEntity, key, item.Pack)
			ce(err)
			newSet = append(newSet, replace...)
			newSet = append(newSet, set[i+1:]...)
			return newSet, nil
		} else {
			panic("impossible")
		}

	} else {
		// append
		set = append(set, SetItem{
			Key: &key,
		})
		return set, nil
	}

}

func mergeKeyToPack(
	fetchEntity entity.Fetch,
	key Key,
	pack *Pack,
) (_ Set, err error) {
	defer he(&err)
	var set Set
	ce(fetchEntity(pack.Key, &set))
	return mergeKeyToSet(fetchEntity, key, set)
}

func mergePackToSet(
	fetchEntity entity.Fetch,
	pack *Pack,
	set Set,
) (_ Set, err error) {
	defer he(&err)

	i := sort.Search(len(set), func(i int) bool {
		item := set[i]
		if item.Key != nil {
			return pack.Max.Compare(*item.Key) <= 0
		} else if item.Pack != nil {
			return pack.Max.Compare(item.Pack.Max) <= 0
		}
		panic("invalid item")
	})

	if i < len(set) {
		// found
		item := set[i]

		if (item.Key != nil && pack.Max.Compare(*item.Key) < 0) ||
			(item.Pack != nil && pack.Max.Compare(item.Pack.Min) < 0) {
			// insert
			newSet := make(Set, 0, len(set)+1)
			newSet = append(newSet, set[:i]...)
			newSet = append(newSet, SetItem{
				Pack: pack,
			})
			newSet = append(newSet, set[i:]...)
			return newSet, nil

		} else if item.Pack != nil && *item.Pack == *pack {
			// same
			return set, nil

		} else if item.Key != nil {
			// merge file
			newSet := make(Set, 0, len(set))
			newSet = append(newSet, set[:i]...)
			replace, err := mergeKeyToPack(fetchEntity, *item.Key, pack)
			ce(err)
			newSet = append(newSet, replace...)
			newSet = append(newSet, set[i+1:]...)
			return newSet, nil

		} else if item.Pack != nil {
			// merge subs
			newSet := make(Set, 0, len(set))
			newSet = append(newSet, set[:i]...)
			replace, err := mergePack(fetchEntity, pack, item.Pack)
			ce(err)
			newSet = append(newSet, replace...)
			newSet = append(newSet, set[i+1:]...)
			return newSet, nil

		} else {
			panic("impossible")
		}

	} else {
		// append
		set = append(set, SetItem{
			Pack: pack,
		})
		return set, nil
	}

}

func mergePack(
	fetchEntity entity.Fetch,
	a *Pack,
	b *Pack,
) (_ Set, err error) {
	defer he(&err)
	var setA Set
	ce(fetchEntity(a.Key, &setA))
	var setB Set
	ce(fetchEntity(b.Key, &setB))
	for _, item := range setA {
		if item.Key != nil {
			setB, err = mergeKeyToSet(fetchEntity, *item.Key, setB)
			ce(err)
		} else if item.Pack != nil {
			setB, err = mergePackToSet(fetchEntity, item.Pack, setB)
			ce(err)
		}
	}
	return setB, nil
}

type Partition struct {
	Begin  int
	End    int
	Height int
	Weight int
}

func pack(
	set Set,
	threshold int,
	saveEntity entity.SaveEntity,
) (_ Set, err error) {
	defer he(&err)

l1:

	// partition by height
	var partitions []*Partition
	for i, item := range set {
		var height int
		var weight int
		if item.Key != nil {
			height = 1
			weight = 1
		}
		if item.Pack != nil {
			height = item.Pack.Height
			weight = 1
		}
		if i == 0 {
			partitions = append(partitions, &Partition{
				Begin:  i,
				End:    i,
				Height: height,
				Weight: weight,
			})
		} else {
			p := partitions[len(partitions)-1]
			if height != p.Height {
				partitions = append(partitions, &Partition{
					Begin:  i,
					End:    i,
					Height: height,
					Weight: weight,
				})
			} else {
				if p.Weight+weight <= threshold {
					p.End = i
					p.Weight += weight
				} else {
					partitions = append(partitions, &Partition{
						Begin:  i,
						End:    i,
						Height: height,
						Weight: weight,
					})
				}
			}
		}
	}

	// find heaviest partition
	sort.Slice(partitions, func(i, j int) bool {
		return partitions[i].Weight > partitions[j].Weight
	})
	partition := partitions[0]
	if partition.Weight < threshold {
		return set, nil
	}

	// pack elements in partition
	var min, max Key
	for i := partition.Begin; i <= partition.End; i++ {
		item := set[i]
		if item.Key != nil {
			if !min.Valid() ||
				item.Key.Compare(min) < 0 {
				min = *item.Key
			}
			if !max.Valid() ||
				item.Key.Compare(max) > 0 {
				max = *item.Key
			}
		}
		if item.Pack != nil {
			if !min.Valid() ||
				item.Pack.Min.Compare(min) < 0 {
				min = item.Pack.Min
			}
			if !max.Valid() ||
				item.Pack.Max.Compare(max) > 0 {
				max = item.Pack.Max
			}
		}
	}
	slice := set[partition.Begin : partition.End+1]

	// pack
	summary, err := saveEntity(slice)
	ce(err)
	replace := SetItem{
		Pack: &Pack{
			Key:    summary.Key,
			Min:    min,
			Max:    max,
			Height: partition.Height + 1,
		},
	}

	// replace
	newSet := make(Set, 0, len(set))
	newSet = append(newSet, set[:partition.Begin]...)
	newSet = append(newSet, replace)
	newSet = append(newSet, set[partition.End+1:]...)
	set = newSet

	// repeat
	goto l1
}

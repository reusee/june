package keyset

import (
	"sort"

	"github.com/reusee/june/entity"
)

type PackThreshold int

func (_ Def) PackThreshold() PackThreshold {
	return 48
}

type PackSet func(
	set Set,
) (
	Set,
	error,
)

func (_ Def) PackSet(
	threshold PackThreshold,
	saveEntity entity.SaveEntity,
) PackSet {

	thresholdWeight := int(threshold)

	return func(
		set Set,
	) (
		newSet Set,
		err error,
	) {
		defer he(&err)

		if len(set) == 0 {
			return set, nil
		}

	l1:

		// partition by height
		var partitions []*Partition
		for i, item := range set {
			var height int
			var weight int
			if item.Key != nil {
				height = 1
				weight = 1
			} else if item.Pack != nil {
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
					if p.Weight+weight <= thresholdWeight {
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
		if partition.Weight < thresholdWeight {
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
			} else if item.Pack != nil {
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
		newSet = make(Set, 0, len(set))
		newSet = append(newSet, set[:partition.Begin]...)
		newSet = append(newSet, replace)
		newSet = append(newSet, set[partition.End+1:]...)
		set = newSet

		// repeat
		goto l1

		return
	}
}

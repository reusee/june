// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package file

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/reusee/june/entity"
	"github.com/reusee/june/sys"
	"github.com/reusee/pr2"
)

type BuildOption interface {
	IsBuildOption()
}

type Build func(
	ctx context.Context,
	root *File,
	cont Sink,
	options ...BuildOption,
) Sink

type SmallFileThreshold int64

type PackThreshold int

func (Def) Build(
	toContents ToContents,
	smallFileThreshold SmallFileThreshold,
	fetchEntity entity.Fetch,
	saveEntity entity.SaveEntity,
	packThreshold PackThreshold,
	parallel sys.Parallel,
	scope Scope,
	save entity.SaveEntity,
) Build {

	// if a partition's weight is larger than threshold, the partition will be packed
	threshold := int(packThreshold)

	return func(
		ctx context.Context,
		root *File,
		cont Sink,
		options ...BuildOption,
	) Sink {

		// reset
		*root = File{
			IsDir: true,
		}

		// options
		var tapBuild []TapBuildFile
		var tapRead []TapReadFile
		for _, option := range options {
			switch option := option.(type) {
			case TapBuildFile:
				tapBuild = append(tapBuild, option)
			case TapReadFile:
				tapRead = append(tapRead, option)
			default: // NOCOVER
				panic(fmt.Errorf("bad option: %T", option))
			}
		}

		// stack
		type StackItem struct {
			File *File
			Path string
		}
		stack := []StackItem{
			{
				Path: ".",
				File: root,
			},
		}

		// async jobs
		wg := pr2.NewWaitGroup(ctx)
		putFn, wait := pr2.Consume(
			wg,
			int(parallel),
			func(_ int, v any) error {
				return v.(func() error)()
			},
			pr2.BacklogSize(parallel*2),
		)

		// stack unwind
		unwind := func() (err error) {
			defer he(&err)
			if len(stack) < 2 { // NOCOVER
				panic("impossible")
			}
			top := stack[len(stack)-1]
			topParent := stack[len(stack)-2]
			// merge file to parent
			subs, err := mergeFileToSubs(
				fetchEntity,
				topParent.File.Subs,
				top.File,
			)
			ce(err)
			topParent.File.Subs = subs
			// pack parent subs
			if topParent.File != root {
				packed, err := pack(ctx, scope, topParent.File.Subs, threshold, save, wait)
				ce(err)
				topParent.File.Subs = packed
			}
			// pop
			stack = stack[:len(stack)-1]
			return nil
		}

		var sink Sink
		sink = func(v any) (_ Sink, err error) {
			defer he(&err)

			// end of stream
			if v == nil {

				// wait async jobs complete
				ce(wait(false))
				ce(wait(true))

				// unwind until root file
				for len(stack) != 1 {
					ce(unwind())
				}

				wg.Cancel()

				return cont, nil
			}

		l1:
			switch value := v.(type) {

			case FileInfoThunk:
				// expand
				value.Expand(true)
				v = value.FileInfo
				goto l1

			case FileInfo:

				// unwind
				parentDir := filepath.Dir(value.Path)
				for stack[len(stack)-1].Path != parentDir {
					ce(unwind())
				}

				// build *File
				var file *File
				if f, ok := value.FileLike.(*File); ok {
					file = f
				} else {
					file = &File{
						IsDir:   value.GetIsDir(scope),
						Name:    value.GetName(scope),
						Size:    value.GetSize(scope),
						Mode:    value.GetMode(scope),
						ModTime: value.GetModTime(scope),
					}
				}
				for _, fn := range tapBuild {
					fn(value, file)
				}

				if file.IsDir {
					// push to stack
					stack = append(stack, StackItem{
						Path: value.Path,
						File: file,
					})

				} else {
					// set Contents
					if file.Size > 0 && len(file.Contents) == 0 && len(file.ContentBytes) == 0 {
						for _, fn := range tapRead {
							fn(value)
						}
						putFn(func() error {
							if err := value.WithReader(scope, func(r io.Reader) (err error) {
								defer he(&err)
								if file.Size <= int64(smallFileThreshold) {
									bs, err := io.ReadAll(r)
									ce(err)
									file.ContentBytes = bs
								} else {
									keys, lengths, err := toContents(ctx, r, file.Size)
									ce(err)
									file.Contents = keys
									file.ChunkLengths = lengths
								}
								return nil
							}); err != nil { // NOCOVER
								if is(err, os.ErrNotExist) || is(err, os.ErrPermission) { // NOCOVER
									// ignore not exists and no permission errors
									return nil
								}
								return err // NOCOVER
							}
							return nil
						})

					}

					// add to parent
					parent := stack[len(stack)-1].File
					subs, err := mergeFileToSubs(
						fetchEntity,
						parent.Subs,
						file,
					)
					ce(err)
					parent.Subs = subs

					// pack parent
					if parent != root {
						packed, err := pack(ctx, scope, parent.Subs, threshold, save, wait)
						ce(err)
						parent.Subs = packed
					}

				}

			case PackThunk:

				// merge manually
				value.Expand(false)

				// unwind
				for stack[len(stack)-1].Path != value.Path {
					ce(unwind())
				}

				// add to parent
				parent := stack[len(stack)-1].File
				subs, err := mergePackToSubs(
					fetchEntity,
					parent.Subs,
					&value.Pack,
				)
				ce(err)
				parent.Subs = subs

				// pack parent
				if parent != root {
					packed, err := pack(ctx, scope, parent.Subs, threshold, save, wait)
					ce(err)
					parent.Subs = packed
				}

			default: // NOCOVER
				return nil, fmt.Errorf("unknown type: %#v", v)
			}

			return sink, nil
		}

		return sink
	}

}

func mergeFileToSubs(
	fetchEntity entity.Fetch,
	subs Subs,
	file *File,
) (_ Subs, err error) {
	defer he(&err)

	i := sort.Search(len(subs), func(i int) bool {
		sub := subs[i]
		if sub.File != nil {
			return file.Name <= sub.File.Name
		} else if sub.Pack != nil {
			return file.Name <= sub.Pack.Max
		}
		panic("invalid sub") // NOCOVER
	})
	if i < len(subs) {
		// found
		sub := subs[i]
		if (sub.File != nil && file.Name < sub.File.Name) ||
			(sub.Pack != nil && file.Name < sub.Pack.Min) {
			// insert
			newSubs := make(Subs, 0, len(subs)+1)
			newSubs = append(newSubs, subs[:i]...)
			newSubs = append(newSubs, Sub{
				File: file,
			})
			newSubs = append(newSubs, subs[i:]...)
			return newSubs, nil
		} else if sub.File != nil && file.Name == sub.File.Name {
			// replace
			subs[i] = Sub{
				File: file,
			}
			return subs, nil
		} else if sub.Pack != nil && file.Name >= sub.Pack.Min && file.Name <= sub.Pack.Max {
			// merge pack
			newSubs := make(Subs, 0, len(subs))
			newSubs = append(newSubs, subs[:i]...)
			replace, err := mergeFileToPack(fetchEntity, sub.Pack, file)
			ce(err)
			newSubs = append(newSubs, replace...)
			newSubs = append(newSubs, subs[i+1:]...)
			return newSubs, nil
		} else {
			panic("impossible")
		}
	} else {
		// append
		subs = append(subs, Sub{
			File: file,
		})
		return subs, nil
	}
}

func mergeFileToPack(
	fetchEntity entity.Fetch,
	pack *Pack,
	file *File,
) (_ Subs, err error) {
	defer he(&err)
	var subs Subs
	ce(fetchEntity(pack.Key, &subs))
	return mergeFileToSubs(fetchEntity, subs, file)
}

func mergePackToSubs(
	fetchEntity entity.Fetch,
	subs Subs,
	pack *Pack,
) (_ Subs, err error) {
	defer he(&err)

	i := sort.Search(len(subs), func(i int) bool {
		sub := subs[i]
		if sub.File != nil {
			return pack.Max <= sub.File.Name
		} else if sub.Pack != nil {
			return pack.Max <= sub.Pack.Max
		}
		panic("invalid sub")
	})
	if i < len(subs) {
		// found
		sub := subs[i]
		if (sub.File != nil && pack.Max < sub.File.Name) ||
			(sub.Pack != nil && pack.Max < sub.Pack.Min) {
			// insert
			newSubs := make(Subs, 0, len(subs)+1)
			newSubs = append(newSubs, subs[:i]...)
			newSubs = append(newSubs, Sub{
				Pack: pack,
			})
			newSubs = append(newSubs, subs[i:]...)
			return newSubs, nil
		} else if sub.Pack != nil && *sub.Pack == *pack {
			// same
			return subs, nil
		} else if sub.File != nil {
			// merge file
			newSubs := make(Subs, 0, len(subs))
			newSubs = append(newSubs, subs[:i]...)
			replace, err := mergeFileToPack(fetchEntity, pack, sub.File)
			ce(err)
			newSubs = append(newSubs, replace...)
			newSubs = append(newSubs, subs[i+1:]...)
			return newSubs, nil
		} else if sub.Pack != nil {
			// merge subs
			newSubs := make(Subs, 0, len(subs))
			newSubs = append(newSubs, subs[:i]...)
			replace, err := mergePack(fetchEntity, pack, sub.Pack)
			ce(err)
			newSubs = append(newSubs, replace...)
			newSubs = append(newSubs, subs[i+1:]...)
			return newSubs, nil
		} else {
			panic("impossible")
		}
	} else {
		// append
		subs = append(subs, Sub{
			Pack: pack,
		})
		return subs, nil
	}
}

func mergePack(
	fetchEntity entity.Fetch,
	a *Pack,
	b *Pack,
) (_ Subs, err error) {
	defer he(&err)

	var subsA Subs
	ce(fetchEntity(a.Key, &subsA))
	var subsB Subs
	ce(fetchEntity(b.Key, &subsB))
	for _, sub := range subsA {
		if sub.File != nil {
			subsB, err = mergeFileToSubs(fetchEntity, subsB, sub.File)
			ce(err)
		} else if sub.Pack != nil {
			subsB, err = mergePackToSubs(fetchEntity, subsB, sub.Pack)
			ce(err)
		}
	}
	return subsB, nil
}

type Partition struct {
	Begin  int
	End    int
	Height int
	Weight int
}

func pack(
	ctx context.Context,
	scope Scope,
	subs Subs,
	threshold int,
	saveEntity entity.SaveEntity,
	wait func(bool) error,
) (_ Subs, err error) {
	defer he(&err)

l1:

	// partition by height
	var partitions []*Partition
	for i, sub := range subs {
		var height int
		var weight int
		if sub.File != nil {
			height = 1
			weight = sub.File.GetWeight(scope)
		}
		if sub.Pack != nil {
			height = sub.Pack.Height
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
				p.End = i
				p.Weight += weight
			}
		}
	}

	// find heaviest partition
	sort.Slice(partitions, func(i, j int) bool {
		return partitions[i].Weight > partitions[j].Weight
	})
	partition := partitions[0]
	if partition.Weight < threshold {
		return subs, nil
	}

	// pack elements in partition
	var min, max string
	for i := partition.Begin; i <= partition.End; i++ {
		sub := subs[i]
		if sub.File != nil {
			if min == "" {
				min = sub.File.Name
			}
			if max == "" {
				max = sub.File.Name
			}
			if sub.File.Name < min {
				min = sub.File.Name
			}
			if sub.File.Name > max {
				max = sub.File.Name
			}
		}
		if sub.Pack != nil {
			if min == "" {
				min = sub.Pack.Min
			}
			if max == "" {
				max = sub.Pack.Max
			}
			if sub.Pack.Min < min {
				min = sub.Pack.Min
			}
			if sub.Pack.Max > max {
				max = sub.Pack.Max
			}
		}
	}
	slice := subs[partition.Begin : partition.End+1]

	// wait potential changes
	ce(wait(false))

	// pack
	summary, err := saveEntity(ctx, slice)
	ce(err)
	var size int64
	for _, sub := range slice {
		if sub.File != nil {
			size += sub.File.GetTreeSize(scope)
		}
		if sub.Pack != nil {
			size += sub.Pack.Size
		}
	}
	replace := Sub{
		Pack: &Pack{
			Size:   size,
			Key:    summary.Key,
			Min:    min,
			Max:    max,
			Height: partition.Height + 1,
		},
	}

	// replace
	newSubs := make([]Sub, 0, len(subs))
	newSubs = append(newSubs, subs[:partition.Begin]...)
	newSubs = append(newSubs, replace)
	newSubs = append(newSubs, subs[partition.End+1:]...)
	subs = newSubs

	// repeat
	goto l1
}

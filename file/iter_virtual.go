// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package file

import (
	"fmt"
	"path/filepath"
	"sort"
)

type IterVirtualOption interface {
	IsIterVirtualOption()
}

type NoSubsSort bool

func (_ NoSubsSort) IsIterVirtualOption() {}

type IterVirtual func(file Virtual, cont Src, options ...IterVirtualOption) Src

func (_ Def) IterVirtual(
	ignore Ignore,
) IterVirtual {

	return func(
		v Virtual,
		cont Src,
		options ...IterVirtualOption,
	) Src {

		var noSubsSort NoSubsSort
		for _, option := range options {
			switch option := option.(type) {
			case NoSubsSort:
				noSubsSort = option
			default:
				panic(fmt.Errorf("unknown option: %T", option))
			}
		}

		var iterFile func(dir string, file Virtual, cont Src) Src
		var iterSubs func(dir string, file Virtual, cont Src) Src

		iterFile = func(dir string, file Virtual, cont Src) Src {
			return func() (*IterItem, Src, error) {
				path := filepath.Join(dir, file.Name)
				if ignore(path, file) {
					return nil, cont, nil
				}
				if file.IsDir {
					next := invalid
					thunk := FileInfoThunk{
						Path: path,
						FileInfo: FileInfo{
							Path:     path,
							FileLike: file,
						},
						Expand: func(expand bool) {
							if expand {
								next = iterSubs(path, file, cont)
							} else {
								next = cont
							}
						},
					}
					return &IterItem{
							FileInfoThunk: &thunk,
						}, func() (*IterItem, Src, error) {
							return nil, next, nil
						}, nil
				} else {
					info := FileInfo{
						FileLike: file,
						Path:     path,
					}
					return &IterItem{
						FileInfo: &info,
					}, cont, nil
				}
			}
		}

		iterSubs = func(path string, file Virtual, cont Src) Src {
			subs := file.Subs
			if !noSubsSort {
				sort.Slice(subs, func(i, j int) bool {
					return subs[i].Name < subs[j].Name
				})
			}
			var src Src
			src = func() (*IterItem, Src, error) {
				if len(subs) == 0 {
					return nil, cont, nil
				}
				sub := subs[0]
				subs = subs[1:]
				return nil, iterFile(path, sub, src), nil
			}
			return src
		}

		return iterFile(".", v, cont)
	}
}

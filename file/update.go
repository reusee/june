// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package file

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/reusee/e4"
	"github.com/reusee/june/fsys"
)

type Update func(
	path string,
	from Src,
	fromTime time.Time,
	src Src,
	watcher *fsys.Watcher,
	options ...UpdateOption,
) Src

type UpdateOption interface {
	IsUpdateOption()
}

//TODO add tombstone file type
//TODO add option to return changeset

func (_ Def) Update(
	zip Zip,
	unzip Unzip,
	scope Scope,
) Update {

	return func(
		path string,
		from Src,
		fromTime time.Time,
		src Src,
		watcher *fsys.Watcher,
		options ...UpdateOption,
	) Src {

		var tapAdd TapAddFileInfo
		var tapDelete TapDeleteFileInfo
		var tapModify TapModifyFileInfo

		for _, option := range options {
			switch option := option.(type) {
			case TapAddFileInfo:
				tapAdd = option
			case TapDeleteFileInfo:
				tapDelete = option
			case TapModifyFileInfo:
				tapModify = option
			default:
				panic(fmt.Errorf("unknown option: %T", option))
			}
		}

		return func() (_ *IterItem, _ Src, err error) {
			defer he(&err, e4.NewInfo("update %s", path))

			dir := filepath.Dir(path)
			selectFile := func(item ZipItem) *IterItem {
				// A is existed file
				// B is new file

				// delete
				if item.B == nil {
					if item.A.FileInfo != nil {
						if tapDelete != nil {
							tapDelete(*item.A.FileInfo)
						}
					} else if item.A.FileInfoThunk != nil {
						item.A.FileInfoThunk.Expand(false)
						if tapDelete != nil {
							tapDelete(item.A.FileInfoThunk.FileInfo)
						}
					} else if item.A.PackThunk != nil {
						item.A.PackThunk.Expand(false)
					}

					return nil
				}

				// item.B is not null from here

				// new
				if item.A == nil {
					if item.B.FileInfo != nil {
						if tapAdd != nil {
							tapAdd(*item.B.FileInfo)
						}
					} else if item.B.FileInfoThunk != nil {
						if tapAdd != nil {
							tapAdd(item.B.FileInfoThunk.FileInfo)
						}
						item.B.FileInfoThunk.Expand(true)
						return &IterItem{
							FileInfo: &item.B.FileInfoThunk.FileInfo,
						}
					} else if item.B.PackThunk != nil {
						item.B.PackThunk.Expand(true)
						return &IterItem{
							Pack: &item.B.PackThunk.Pack,
						}
					}
					return item.B
				}

				// both item.A and item.B is not null from here

				// handle FileInfoThunk
				if item.A.FileInfoThunk != nil {
					if item.B.FileInfoThunk != nil {
						thunkPath := filepath.Join(dir, item.A.FileInfoThunk.Path)
						if watcher != nil {
							notChanged, err := watcher.PathNotChanged(thunkPath, fromTime)
							ce(err)
							if notChanged {
								// not changed
								item.A.FileInfoThunk.Expand(false)
								item.B.FileInfoThunk.Expand(false)
								return &IterItem{
									FileInfo: &item.A.FileInfoThunk.FileInfo,
								}
							}
						}
					}
				}

				if item.A.FileInfoThunk != nil {
					item.A.FileInfoThunk.Expand(true)
					item.A.FileInfo = &item.A.FileInfoThunk.FileInfo
					item.A.FileInfoThunk = nil
				}
				if item.B.FileInfoThunk != nil {
					item.B.FileInfoThunk.Expand(true)
					item.B.FileInfo = &item.B.FileInfoThunk.FileInfo
					item.B.FileInfoThunk = nil
				}

				// both item.A and item.B is not FileInfoThunk from here

				if item.A.FileInfo != nil {
					if item.B.FileInfo != nil {
						infoA := *item.A.FileInfo
						infoB := *item.B.FileInfo
						aIsDir := infoA.GetIsDir(scope)
						bIsDir := infoB.GetIsDir(scope)
						if bIsDir {
							// do not reuse dirs
							if !(aIsDir == bIsDir &&
								infoA.GetName(scope) == infoB.GetName(scope) &&
								infoA.GetSize(scope) == infoB.GetSize(scope) &&
								infoA.GetMode(scope) == infoB.GetMode(scope) &&
								infoA.GetModTime(scope) == infoB.GetModTime(scope)) {
								// changed
								if tapModify != nil {
									tapModify(infoB, infoA)
								}
							}
							return item.B
						}
						if aIsDir == bIsDir &&
							infoA.GetName(scope) == infoB.GetName(scope) &&
							infoA.GetSize(scope) == infoB.GetSize(scope) &&
							infoA.GetMode(scope) == infoB.GetMode(scope) &&
							infoA.GetModTime(scope) == infoB.GetModTime(scope) {
							// not changed
							return item.A
						} else {
							// changed
							if tapModify != nil {
								tapModify(infoB, infoA)
							}
						}
					}
				}

				return item.B
			}

			return nil, unzip(
				zip(
					from,
					src,
					nil,
					PredictExpandFileInfoThunk(func(a, b FileInfoThunk) (bool, error) {
						// expand manually
						return false, nil
					}),
				),
				selectFile,
				nil,
			), nil

		}
	}
}

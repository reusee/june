// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package file

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/reusee/e4"
	"github.com/reusee/ling/v2/fsys"
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

		return func() (_ any, _ Src, err error) {
			defer he(&err, e4.NewInfo("update %s", path))

			dir := filepath.Dir(path)
			selectFile := func(item ZipItem) any {
				// A is existed file
				// B is new file

				// delete
				if item.B == nil {
					switch v := item.A.(type) {
					case FileInfo:
						if tapDelete != nil {
							tapDelete(v)
						}
					case FileInfoThunk:
						v.Expand(false)
						if tapDelete != nil {
							tapDelete(v.FileInfo)
						}
					case PackThunk:
						v.Expand(false)
					}
					return nil
				}

				// item.B is not null from here

				// new
				if item.A == nil {
					switch v := item.B.(type) {
					case FileInfo:
						if tapAdd != nil {
							tapAdd(v)
						}
					case FileInfoThunk:
						if tapAdd != nil {
							tapAdd(v.FileInfo)
						}
						v.Expand(true)
						return v.FileInfo
					case PackThunk:
						v.Expand(true)
						return v.Pack
					}
					return item.B
				}

				// both item.A and item.B is not null from here

				// handle FileInfoThunk
				if thunkA, ok := item.A.(FileInfoThunk); ok {
					if thunkB, ok := item.B.(FileInfoThunk); ok {
						thunkPath := filepath.Join(dir, thunkA.Path)
						if watcher != nil {
							notChanged, err := watcher.PathNotChanged(thunkPath, fromTime)
							ce(err)
							if notChanged {
								// not changed
								thunkA.Expand(false)
								thunkB.Expand(false)
								return thunkA.FileInfo
							}
						}
					}
				}

				if t, ok := item.A.(FileInfoThunk); ok {
					t.Expand(true)
					item.A = t.FileInfo
				}
				if t, ok := item.B.(FileInfoThunk); ok {
					t.Expand(true)
					item.B = t.FileInfo
				}

				// both item.A and item.B is not FileInfoThunk from here

				if infoA, ok := item.A.(FileInfo); ok {
					if infoB, ok := item.B.(FileInfo); ok {
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

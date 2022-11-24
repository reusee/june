// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package file

import (
	"context"
	"path/filepath"

	"github.com/reusee/e5"
	"github.com/reusee/june/entity"
)

type IterFile func(ctx context.Context, file *File, cont Src) Src

type IterKey func(ctx context.Context, key Key, cont Src) Src

func (Def) IterFile(
	fetch entity.Fetch,
	ignore Ignore,
) (
	IterFile,
	IterKey,
) {

	var iterPack func(ctx context.Context, dir string, pack Pack, cont Src) Src
	var iterSubs func(ctx context.Context, dir string, file *File, cont Src) Src
	var iterFile func(ctx context.Context, dir string, file *File, cont Src) Src
	var iterKey func(ctx context.Context, dir string, key Key, cont Src) Src

	iterPack = func(ctx context.Context, path string, pack Pack, cont Src) Src {
		loaded := false
		var subs Subs
		var src Src
		src = func() (_ any, _ Src, err error) {
			defer he(&err)
			if !loaded {
				err := fetch(pack.Key, &subs)
				ce(err, e5.Info("fetch pack %s %s", path, pack.Key))
				loaded = true
			}
			if len(subs) == 0 {
				return nil, cont, nil
			}
			sub := subs[0]
			subs = subs[1:]
			if sub.File != nil {
				return nil, iterFile(ctx, path, sub.File, src), nil
			}
			if sub.Pack != nil {
				next := invalid
				thunk := PackThunk{
					Path: path,
					Pack: *sub.Pack,
					Expand: func(expand bool) {
						if expand {
							next = iterPack(ctx, path, *sub.Pack, src)
						} else {
							next = src
						}
					},
				}
				return thunk, func() (any, Src, error) {
					return nil, next, nil
				}, nil
			}
			return nil, src, nil
		}
		return src
	}

	iterFile = func(ctx context.Context, dir string, file *File, cont Src) Src {
		return func() (any, Src, error) {
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
							next = iterSubs(ctx, path, file, cont)
						} else {
							next = cont
						}
					},
				}
				return thunk, func() (any, Src, error) {
					return nil, next, nil
				}, nil
			} else {
				info := FileInfo{
					FileLike: file,
					Path:     path,
				}
				return info, cont, nil
			}
		}
	}

	iterSubs = func(ctx context.Context, path string, file *File, cont Src) Src {
		subs := file.Subs
		var src Src
		src = func() (any, Src, error) {
			if len(subs) == 0 {
				return nil, cont, nil
			}
			sub := subs[0]
			subs = subs[1:]
			if sub.File != nil {
				return nil, iterFile(ctx, path, sub.File, src), nil
			}
			if sub.Pack != nil {
				next := invalid
				thunk := PackThunk{
					Path: path,
					Pack: *sub.Pack,
					Expand: func(expand bool) {
						if expand {
							next = iterPack(ctx, path, *sub.Pack, src)
						} else {
							next = src
						}
					},
				}
				return thunk, func() (any, Src, error) {
					return nil, next, nil
				}, nil
			}
			return nil, src, nil
		}
		return src
	}

	iterKey = func(ctx context.Context, dir string, key Key, cont Src) Src {
		return func() (_ any, _ Src, err error) {
			defer he(&err)
			var file File
			err = fetch(key, &file)
			ce(err, e5.Info("iter %s %s", dir, key))
			return nil, iterFile(ctx, dir, &file, cont), nil
		}
	}

	return func(ctx context.Context, file *File, cont Src) Src {
			return iterFile(ctx, ".", file, cont)
		},
		func(ctx context.Context, key Key, cont Src) Src {
			return iterKey(ctx, ".", key, cont)
		}
}

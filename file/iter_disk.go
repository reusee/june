// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package file

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
	"github.com/reusee/e5"
	"github.com/reusee/june/fsys"
)

type IterDiskFile func(
	ctx context.Context,
	path string,
	cont Src,
	options ...IterDiskFileOption,
) Src

type IterDiskFileOption interface {
	IsIterDiskFileOption()
}

func (Def) IterDiskFile(
	ignore Ignore,
	isRestrictedPath fsys.IsRestrictedPath,
) (iter IterDiskFile) {

	type Ignore = func(path string, file DiskFile) bool

	type Options struct {
		UseGitIgnore bool
		SingleDevice bool
		Device       uint64
	}

	var iterFile func(
		ctx context.Context,
		base string,
		rel string,
		options Options,
		ignores []Ignore,
		cont Src,
	) Src

	iterSubs := func(
		ctx context.Context,
		base string,
		rel string,
		diskFile DiskFile,
		options Options,
		ignores []Ignore,
		cont Src,
	) Src {

		var src Src
		var file *os.File
		var names []string
		src = func() (_ any, _ Src, err error) {
			filePath := filepath.Join(base, rel)
			defer he(&err, e5.Info("iter subs %s", filePath))

			if file == nil {
				// open file lazily
				var err error
				file, err = os.Open(filePath)
				if is(err, os.ErrNotExist) || is(err, os.ErrPermission) {
					// ignore not exists and no permission errors
					return nil, cont, nil
				}
				ce(err)
				defer file.Close()
				names, err = file.Readdirnames(-1)
				ce(err)
				sort.Slice(names, func(i, j int) bool {
					return names[i] < names[j]
				})

				// handle .gitignore
				if options.UseGitIgnore {
					ignoreFilePath := filepath.Join(filePath, ".gitignore")
					content, err := os.ReadFile(ignoreFilePath)
					if err == nil {
						lines := bytes.FieldsFunc(content, func(r rune) bool {
							return r == '\r' || r == '\n'
						})
						domain := strings.Split(filePath, PathSeparator)
						for _, line := range lines {
							if len(line) == 0 {
								continue
							}
							pattern := gitignore.ParsePattern(string(line), domain)
							ignores = append(ignores, func(path string, file DiskFile) bool {
								return pattern.Match(
									strings.Split(path, PathSeparator),
									file.info.IsDir(),
								) == gitignore.Exclude
							})

						}
					}
				}

			}
			if len(names) == 0 {
				return nil, cont, nil
			}
			name := names[0]
			names = names[1:]
			return nil, iterFile(
				ctx,
				base,
				filepath.Join(rel, name),
				options,
				ignores,
				src,
			), nil
		}
		return src
	}

	iterFile = func(
		ctx context.Context,
		base string,
		rel string,
		options Options,
		ignores []Ignore,
		cont Src,
	) Src {

		return func() (_ any, _ Src, err error) {

			select {
			case <-ctx.Done():
				err = ctx.Err()
				return
			default:
			}

			diskPath := filepath.Join(base, rel)
			defer he(&err, e5.Info("iter file %s", diskPath))

			ok, err := isRestrictedPath(diskPath)
			ce(err)
			if ok {
				return nil, cont, nil
			}
			stat, err := os.Lstat(diskPath)
			if is(err, os.ErrNotExist) || is(err, os.ErrPermission) {
				// ignore not exists and no permission errors
				return nil, cont, nil
			}
			ce(err)
			if options.SingleDevice {
				if fsys.GetDevice(stat) != options.Device {
					return nil, cont, nil
				}
			}
			mode := stat.Mode()
			if mode&(0|
				os.ModeNamedPipe|os.ModeSocket|
				os.ModeDevice|os.ModeCharDevice|
				os.ModeIrregular) > 0 {
				return nil, cont, nil
			}
			diskFile := DiskFile{
				info: stat,
				Path: diskPath,
			}

			if mode&os.ModeSymlink > 0 {
				diskFile.contentReady = true
				s, err := os.Readlink(filepath.Join(base, rel))
				ce(err)
				diskFile.content = []byte(s)
			}

			if ignore(rel, diskFile) {
				return nil, cont, nil
			}
			for _, fn := range ignores {
				if ok := fn(filepath.Join(base, rel), diskFile); ok {
					return nil, cont, nil
				}
			}
			if diskFile.info.IsDir() {
				next := invalid
				thunk := FileInfoThunk{
					Path: rel,
					FileInfo: FileInfo{
						Path:     rel,
						FileLike: diskFile,
					},
					Expand: func(expand bool) {
						if expand {
							next = iterSubs(ctx, base, rel, diskFile, options, ignores, cont)
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
					Path:     rel,
					FileLike: diskFile,
				}
				return info, cont, nil
			}
		}
	}

	return func(ctx context.Context, path string, cont Src, options ...IterDiskFileOption) Src {
		return func() (_ any, _ Src, err error) {
			defer he(&err, e5.Info("iter %s", path))

			abs, err := fsys.RealPath(path)
			ce(err)

			var opts Options
			for _, option := range options {
				switch option := option.(type) {
				case UseGitIgnore:
					opts.UseGitIgnore = bool(option)
				case SingleDevice:
					opts.SingleDevice = bool(option)
					if option {
						stat, err := os.Lstat(abs)
						ce(err)
						opts.Device = fsys.GetDevice(stat)
					}
				default:
					panic(fmt.Errorf("unknown option: %T", option))
				}
			}

			base, rel := filepath.Split(abs)
			return nil, iterFile(
				ctx,
				base,
				rel,
				opts,
				nil,
				cont,
			), nil
		}
	}

}

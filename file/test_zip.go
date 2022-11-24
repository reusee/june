// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package file

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/reusee/e5"
	"github.com/reusee/june/fsys"
	"github.com/reusee/pp"
	"github.com/reusee/pr2"
)

func TestZip(
	t *testing.T,
	scope Scope,
) {
	defer he(nil, e5.TestingFatal(t))

	scope.Fork(
		func() PackThreshold {
			return 2
		},
	).Call(func(
		zip Zip,
		iterVirtual IterVirtual,
		build Build,
		unzip Unzip,
		equal Equal,
		reverse Reverse,
		iterFile IterFile,
		wg *pr2.WaitGroup,
	) {

		predictExpand := PredictExpandFileInfoThunk(func(a, b FileInfoThunk) (bool, error) {
			return !a.FileInfo.GetModTime(scope).Equal(b.FileInfo.GetModTime(scope)), nil
		})

		D := func(name string, subs ...Virtual) Virtual {
			return Virtual{
				IsDir: true,
				Name:  name,
				Subs:  subs,
			}
		}
		DT := func(name string, modtime time.Time, subs ...Virtual) Virtual {
			return Virtual{
				IsDir:   true,
				Name:    name,
				ModTime: modtime,
				Subs:    subs,
			}
		}
		F := func(name string) Virtual {
			return Virtual{
				Name: name,
			}
		}
		I := func(path string, v Virtual) FileInfo {
			return FileInfo{
				Path:     path,
				FileLike: v,
			}
		}
		T := func(path string, v Virtual) FileInfoThunk {
			return FileInfoThunk{
				Path: path,
				FileInfo: FileInfo{
					Path:     path,
					FileLike: v,
				},
				Expand: func(bool) {
					// ignore
				},
			}
		}

		file1 := new(File)
		err := pp.Copy(
			iterVirtual(D("foo", D("bar", D("baz", D("qux", F("quux"))))), nil),
			build(wg, file1, nil),
		)
		ce(err)
		file1 = file1.Subs[0].File
		file2 := new(File)
		err = pp.Copy(
			iterVirtual(D("foo", D("1", D("baz", D("qux", F("quux"))))), nil),
			build(wg, file2, nil),
		)
		ce(err)
		file2 = file2.Subs[0].File

		type Case struct {
			A        any
			B        any
			Expected []ZipItem
			Skip     bool
		}
		cases := []Case{

			// FileInfo
			0: {
				A: F("foo"),
				B: F("foo"),
				Expected: []ZipItem{
					{
						A:   I("foo", F("foo")),
						B:   I("foo", F("foo")),
						Dir: ".",
					},
				},
			},

			1: {
				A: F("bar"),
				B: F("foo"),
				Expected: []ZipItem{
					{
						A:   I("bar", F("bar")),
						B:   nil,
						Dir: ".",
					},
					{
						A:   nil,
						B:   I("foo", F("foo")),
						Dir: ".",
					},
				},
			},

			2: {
				A: F("foo"),
				B: F("bar"),
				Expected: []ZipItem{
					{
						A:   nil,
						B:   I("bar", F("bar")),
						Dir: ".",
					},
					{
						A:   I("foo", F("foo")),
						B:   nil,
						Dir: ".",
					},
				},
			},

			// TreeInfo
			3: {
				A: D("foo",
					F("foo"),
				),
				B: D("foo",
					F("foo"),
				),
				Expected: []ZipItem{
					{
						A:   T("foo", D("foo")),
						B:   T("foo", D("foo")),
						Dir: ".",
					},
				},
			},

			4: {
				A: D("bar",
					F("bar"),
				),
				B: D("foo",
					F("foo"),
				),
				Expected: []ZipItem{
					{
						A:   T("bar", D("bar")),
						B:   nil,
						Dir: ".",
					},
					{
						A:   nil,
						B:   T("foo", D("foo")),
						Dir: ".",
					},
				},
			},

			5: {
				A: D("foo",
					F("foo"),
				),
				B: D("bar",
					F("bar"),
				),
				Expected: []ZipItem{
					{
						A:   nil,
						B:   T("bar", D("bar")),
						Dir: ".",
					},
					{
						A:   T("foo", D("foo")),
						B:   nil,
						Dir: ".",
					},
				},
			},

			// TreeInfo and FileInfo
			6: {
				A: D("foo", F("foo")),
				B: ExpandAll(iterVirtual(D("foo", F("foo")), nil)),
				Expected: []ZipItem{
					{
						A:   I("foo", D("foo")),
						B:   I("foo", D("foo")),
						Dir: ".",
					},
					{
						A:   I("foo/foo", D("foo")),
						B:   I("foo/foo", D("foo")),
						Dir: "foo",
					},
				},
			},

			7: {
				A: D("foo", F("bar")),
				B: ExpandAll(iterVirtual(D("foo", F("foo")), nil)),
				Expected: []ZipItem{
					{
						A:   I("foo", D("foo")),
						B:   I("foo", D("foo")),
						Dir: ".",
					},
					{
						A:   I("foo/bar", D("bar")),
						B:   nil,
						Dir: "foo",
					},
					{
						A:   nil,
						B:   I("foo/foo", D("foo")),
						Dir: "foo",
					},
				},
			},

			8: {
				A: ExpandAll(iterVirtual(D("foo", F("foo")), nil)),
				B: D("foo", F("bar")),
				Expected: []ZipItem{
					{
						A:   I("foo", D("foo")),
						B:   I("foo", D("foo")),
						Dir: ".",
					},
					{
						A:   nil,
						B:   I("foo/bar", D("bar")),
						Dir: "foo",
					},
					{
						A:   I("foo/foo", D("foo")),
						B:   nil,
						Dir: "foo",
					},
				},
			},

			// TreeInfo and modify time
			9: {
				A: DT("foo",
					time.Date(2000, 1, 1, 1, 1, 1, 1, time.Local),
					F("foo"),
				),
				B: D("foo",
					F("foo"),
				),
				Expected: []ZipItem{
					{
						A:   I("foo", D("foo")),
						B:   I("foo", D("foo")),
						Dir: ".",
					},
					{
						A:   I("foo/foo", F("foo")),
						B:   I("foo/foo", F("foo")),
						Dir: "foo",
					},
				},
			},

			// dir and file
			10: {
				A: D("foo"),
				B: F("foo"),
				Expected: []ZipItem{
					{
						A:   I("foo", D("foo")),
						B:   I("foo", F("foo")),
						Dir: ".",
					},
				},
			},

			11: {
				A: F("foo"),
				B: D("foo"),
				Expected: []ZipItem{
					{
						A:   I("foo", F("foo")),
						B:   I("foo", D("foo")),
						Dir: ".",
					},
				},
			},

			12: {
				A: F("foo"),
				B: D("foo", F("foo")),
				Expected: []ZipItem{
					{
						A:   I("foo", F("foo")),
						B:   I("foo", D("foo")),
						Dir: ".",
					},
					{
						A:   nil,
						B:   I("foo/foo", F("foo")),
						Dir: "foo",
					},
				},
			},

			13: {
				A: D("foo", F("foo")),
				B: F("foo"),
				Expected: []ZipItem{
					{
						A:   I("foo", F("foo")),
						B:   I("foo", D("foo")),
						Dir: ".",
					},
					{
						A:   I("foo/foo", F("foo")),
						B:   nil,
						Dir: "foo",
					},
				},
			},

			14: {
				A: ExpandAll(
					iterVirtual(
						D("foo", D("bar", F("baz"))),
						nil,
					),
				),
				B: D("foo"),
				Expected: []ZipItem{
					{
						A:   I("foo", D("foo")),
						B:   I("foo", D("foo")),
						Dir: ".",
					},
					{
						A:   I("foo/bar", D("bar")),
						B:   nil,
						Dir: "foo",
					},
					{
						A:   I("foo/bar/baz", D("baz")),
						B:   nil,
						Dir: "foo/bar",
					},
				},
			},

			15: {
				A: D("foo"),
				B: ExpandAll(
					iterVirtual(
						D("foo", D("bar", F("baz"))),
						nil,
					),
				),
				Expected: []ZipItem{
					{
						A:   I("foo", D("foo")),
						B:   I("foo", D("foo")),
						Dir: ".",
					},
					{
						A:   nil,
						B:   I("foo/bar", D("bar")),
						Dir: "foo",
					},
					{
						A:   nil,
						B:   I("foo/bar/baz", F("baz")),
						Dir: "foo/bar",
					},
				},
			},

			16: {
				A: D("bar"),
				B: ExpandAll(
					iterVirtual(
						D("foo", D("bar", F("baz"))),
						nil,
					),
				),
				Expected: []ZipItem{
					{
						A:   T("bar", D("bar")),
						B:   nil,
						Dir: ".",
					},
					{
						A:   nil,
						B:   I("foo", D("foo")),
						Dir: ".",
					},
					{
						A:   nil,
						B:   I("foo/bar", D("bar")),
						Dir: "foo",
					},
					{
						A:   nil,
						B:   I("foo/bar/baz", F("baz")),
						Dir: "foo/bar",
					},
				},
			},

			17: {
				A: ExpandAll(
					iterVirtual(
						D("foo", D("bar", F("baz"))),
						nil,
					),
				),
				B: D("bar"),
				Expected: []ZipItem{
					{
						A:   nil,
						B:   T("bar", D("bar")),
						Dir: ".",
					},
					{
						A:   I("foo", D("foo")),
						B:   nil,
						Dir: ".",
					},
					{
						A:   I("foo/bar", D("bar")),
						B:   nil,
						Dir: "foo",
					},
					{
						A:   I("foo/bar/baz", F("baz")),
						B:   nil,
						Dir: "foo/bar",
					},
				},
			},

			18: {
				A: ExpandAll(
					iterVirtual(
						D("1", D("11", F("111"))),
						nil,
					),
				),
				B: ExpandAll(
					iterVirtual(
						D("2", D("22", F("222"))),
						nil,
					),
				),
				Expected: []ZipItem{
					{
						A:   I("1", D("1")),
						B:   nil,
						Dir: ".",
					},
					{
						A:   I("1/11", D("11")),
						B:   nil,
						Dir: "1",
					},
					{
						A:   I("1/11/111", F("111")),
						B:   nil,
						Dir: "1/11",
					},
					{
						A:   nil,
						B:   I("2", D("2")),
						Dir: ".",
					},
					{
						A:   nil,
						B:   I("2/22", D("22")),
						Dir: "2",
					},
					{
						A:   nil,
						B:   I("2/22/222", F("222")),
						Dir: "2/22",
					},
				},
			},

			19: {
				A: ExpandAll(
					iterVirtual(
						D("2", D("22", F("222"))),
						nil,
					),
				),
				B: ExpandAll(
					iterVirtual(
						D("1", D("11", F("111"))),
						nil,
					),
				),
				Expected: []ZipItem{
					{
						A:   nil,
						B:   I("1", D("1")),
						Dir: ".",
					},
					{
						A:   nil,
						B:   I("1/11", D("11")),
						Dir: "1",
					},
					{
						A:   nil,
						B:   I("1/11/111", F("111")),
						Dir: "1/11",
					},
					{
						A:   I("2", D("2")),
						B:   nil,
						Dir: ".",
					},
					{
						A:   I("2/22", D("22")),
						B:   nil,
						Dir: "2",
					},
					{
						A:   I("2/22/222", F("222")),
						B:   nil,
						Dir: "2/22",
					},
				},
			},

			20: {
				A: D("bar"),
				B: ExpandAll(
					iterVirtual(
						D("1", D("bar", F("baz"))),
						nil,
					),
				),
				Expected: []ZipItem{
					{
						A:   nil,
						B:   I("1", D("1")),
						Dir: ".",
					},
					{
						A:   nil,
						B:   I("1/bar", D("bar")),
						Dir: "1",
					},
					{
						A:   nil,
						B:   I("1/bar/baz", F("baz")),
						Dir: "1/bar",
					},
					{
						A:   T("bar", D("bar")),
						B:   nil,
						Dir: ".",
					},
				},
			},

			21: {
				B: D("bar"),
				A: ExpandAll(
					iterVirtual(
						D("1", D("bar", F("baz"))),
						nil,
					),
				),
				Expected: []ZipItem{
					{
						A:   I("1", D("1")),
						B:   nil,
						Dir: ".",
					},
					{
						A:   I("1/bar", D("bar")),
						B:   nil,
						Dir: "1",
					},
					{
						A:   I("1/bar/baz", F("baz")),
						B:   nil,
						Dir: "1/bar",
					},
					{
						A:   nil,
						B:   T("bar", D("bar")),
						Dir: ".",
					},
				},
			},

			22: {
				A: pp.Seq(
					I("foo", D("foo")),
					T("foo/bar", D("bar")),
				),
				B: pp.Seq(
					I("foo", D("foo")),
				),
				Expected: []ZipItem{
					{
						A:   I("foo", D("foo")),
						B:   I("foo", D("foo")),
						Dir: ".",
					},
					{
						A:   T("foo/bar", D("bar")),
						B:   nil,
						Dir: "foo",
					},
				},
			},

			23: {
				A: pp.Seq(
					I("foo", D("foo")),
				),
				B: pp.Seq(
					I("foo", D("foo")),
					T("foo/bar", D("bar")),
				),
				Expected: []ZipItem{
					{
						A:   I("foo", D("foo")),
						B:   I("foo", D("foo")),
						Dir: ".",
					},
					{
						A:   nil,
						B:   T("foo/bar", D("bar")),
						Dir: "foo",
					},
				},
			},

			24: {
				A: iterFile(file1, nil),
				B: ExpandAll(
					iterVirtual(D("foo"), nil),
				),
				Expected: []ZipItem{
					{
						A:   I("foo", D("foo")),
						B:   I("foo", D("foo")),
						Dir: ".",
					},
					{
						A: PackThunk{
							Path: "foo",
						},
						B:   nil,
						Dir: "foo",
					},
				},
			},

			25: {
				A: ExpandAll(
					iterVirtual(D("foo"), nil),
				),
				B: iterFile(file1, nil),
				Expected: []ZipItem{
					{
						A:   I("foo", D("foo")),
						B:   I("foo", D("foo")),
						Dir: ".",
					},
					{
						A: nil,
						B: PackThunk{
							Path: "foo",
						},
						Dir: "foo",
					},
				},
			},

			26: {
				A: ExpandAll(
					iterVirtual(D("foo", D("bar")), nil),
				),
				B: iterFile(file1, nil),
				Expected: []ZipItem{
					{
						A:   I("foo", D("foo")),
						B:   I("foo", D("foo")),
						Dir: ".",
					},
					{
						A:   I("foo/bar", D("bar")),
						B:   I("foo/bar", D("bar")),
						Dir: "foo",
					},
					{
						A:   nil,
						B:   PackThunk{Path: "foo/bar"},
						Dir: "foo/bar",
					},
				},
			},

			27: {
				A: iterFile(file1, nil),
				B: ExpandAll(
					iterVirtual(D("foo", D("bar")), nil),
				),
				Expected: []ZipItem{
					{
						A:   I("foo", D("foo")),
						B:   I("foo", D("foo")),
						Dir: ".",
					},
					{
						A:   I("foo/bar", D("bar")),
						B:   I("foo/bar", D("bar")),
						Dir: "foo",
					},
					{
						A:   PackThunk{Path: "foo/bar"},
						B:   nil,
						Dir: "foo/bar",
					},
				},
			},

			28: {
				A: iterFile(file1, nil),
				B: ExpandAll(
					iterVirtual(D("foo", D("1")), nil),
				),
				Expected: []ZipItem{
					{
						A:   I("foo", D("foo")),
						B:   I("foo", D("foo")),
						Dir: ".",
					},
					{
						A:   nil,
						B:   I("foo/1", D("1")),
						Dir: "foo",
					},
					{
						A:   PackThunk{Path: "foo"},
						B:   nil,
						Dir: "foo",
					},
				},
			},

			29: {
				A: ExpandAll(
					iterVirtual(D("foo", D("1")), nil),
				),
				B: iterFile(file1, nil),
				Expected: []ZipItem{
					{
						A:   I("foo", D("foo")),
						B:   I("foo", D("foo")),
						Dir: ".",
					},
					{
						A:   I("foo/1", D("1")),
						B:   nil,
						Dir: "foo",
					},
					{
						A:   nil,
						B:   PackThunk{Path: "foo"},
						Dir: "foo",
					},
				},
			},

			30: {
				A: ExpandAll(
					iterVirtual(D("foo", D("z")), nil),
				),
				B: iterFile(file1, nil),
				Expected: []ZipItem{
					{
						A:   I("foo", D("foo")),
						B:   I("foo", D("foo")),
						Dir: ".",
					},
					{
						A:   nil,
						B:   PackThunk{Path: "foo"},
						Dir: "foo",
					},
					{
						A:   I("foo/z", D("z")),
						B:   nil,
						Dir: "foo",
					},
				},
			},

			31: {
				A: iterFile(file1, nil),
				B: ExpandAll(
					iterVirtual(D("foo", D("z")), nil),
				),
				Expected: []ZipItem{
					{
						A:   I("foo", D("foo")),
						B:   I("foo", D("foo")),
						Dir: ".",
					},
					{
						A:   PackThunk{Path: "foo"},
						B:   nil,
						Dir: "foo",
					},
					{
						A:   nil,
						B:   I("foo/z", D("z")),
						Dir: "foo",
					},
				},
			},

			32: {
				A: iterFile(file1, nil),
				B: pp.Seq(
					I("foo", D("foo")),
					T("foo/bar", D("bar")),
				),
				Expected: []ZipItem{
					{
						A:   I("foo", D("foo")),
						B:   I("foo", D("foo")),
						Dir: ".",
					},
					{
						A:   T("foo/bar", D("bar")),
						B:   T("foo/bar", D("bar")),
						Dir: "foo",
					},
				},
			},

			33: {
				A: pp.Seq(
					I("foo", D("foo")),
					T("foo/bar", D("bar")),
				),
				B: iterFile(file1, nil),
				Expected: []ZipItem{
					{
						A:   I("foo", D("foo")),
						B:   I("foo", D("foo")),
						Dir: ".",
					},
					{
						A:   T("foo/bar", D("bar")),
						B:   T("foo/bar", D("bar")),
						Dir: "foo",
					},
				},
			},

			34: {
				A: pp.Seq(
					I("foo", D("foo")),
					T("foo/1", D("1")),
				),
				B: iterFile(file1, nil),
				Expected: []ZipItem{
					{
						A:   I("foo", D("foo")),
						B:   I("foo", D("foo")),
						Dir: ".",
					},
					{
						A:   T("foo/1", D("1")),
						B:   nil,
						Dir: "foo",
					},
					{
						A:   nil,
						B:   PackThunk{Path: "foo"},
						Dir: "foo",
					},
				},
			},

			35: {
				A: iterFile(file1, nil),
				B: pp.Seq(
					I("foo", D("foo")),
					T("foo/1", D("1")),
				),
				Expected: []ZipItem{
					{
						A:   I("foo", D("foo")),
						B:   I("foo", D("foo")),
						Dir: ".",
					},
					{
						A:   nil,
						B:   T("foo/1", D("1")),
						Dir: "foo",
					},
					{
						A:   PackThunk{Path: "foo"},
						B:   nil,
						Dir: "foo",
					},
				},
			},

			36: {
				A: iterFile(file1, nil),
				B: pp.Seq(
					I("foo", D("foo")),
					T("foo/z", D("z")),
				),
				Expected: []ZipItem{
					{
						A:   I("foo", D("foo")),
						B:   I("foo", D("foo")),
						Dir: ".",
					},
					{
						A:   PackThunk{Path: "foo"},
						B:   nil,
						Dir: "foo",
					},
					{
						A:   nil,
						B:   T("foo/z", D("z")),
						Dir: "foo",
					},
				},
			},

			37: {
				A: pp.Seq(
					I("foo", D("foo")),
					T("foo/z", D("z")),
				),
				B: iterFile(file1, nil),
				Expected: []ZipItem{
					{
						A:   I("foo", D("foo")),
						B:   I("foo", D("foo")),
						Dir: ".",
					},
					{
						A:   nil,
						B:   PackThunk{Path: "foo"},
						Dir: "foo",
					},
					{
						A:   T("foo/z", D("z")),
						B:   nil,
						Dir: "foo",
					},
				},
			},

			38: {
				A: ExpandAll(iterFile(file1, nil)),
				B: ExpandAll(iterFile(file1, nil)),
				Expected: []ZipItem{
					{
						A:   I("foo", D("foo")),
						B:   I("foo", D("foo")),
						Dir: ".",
					},
					{
						A:   I("foo/bar", D("bar")),
						B:   I("foo/bar", D("bar")),
						Dir: "foo",
					},
					{
						A:   I("foo/bar/baz", D("baz")),
						B:   I("foo/bar/baz", D("baz")),
						Dir: "foo/bar",
					},
					{
						A:   I("foo/bar/baz/qux", D("qux")),
						B:   I("foo/bar/baz/qux", D("qux")),
						Dir: "foo/bar/baz",
					},
					{
						A:   I("foo/bar/baz/qux/quux", F("quux")),
						B:   I("foo/bar/baz/qux/quux", F("quux")),
						Dir: "foo/bar/baz/qux",
					},
				},
			},

			39: {
				A: ExpandFileInfoThunk(iterFile(file1, nil)),
				B: ExpandFileInfoThunk(iterFile(file2, nil)),
				Expected: []ZipItem{
					{
						A:   I("foo", D("foo")),
						B:   I("foo", D("foo")),
						Dir: ".",
					},
					{
						A:   nil,
						B:   PackThunk{Path: "foo"},
						Dir: "foo",
					},
					{
						A:   PackThunk{Path: "foo"},
						B:   nil,
						Dir: "foo",
					},
				},
			},

			40: {
				A: ExpandFileInfoThunk(iterFile(file2, nil)),
				B: ExpandFileInfoThunk(iterFile(file1, nil)),
				Expected: []ZipItem{
					{
						A:   I("foo", D("foo")),
						B:   I("foo", D("foo")),
						Dir: ".",
					},
					{
						A:   PackThunk{Path: "foo"},
						B:   nil,
						Dir: "foo",
					},
					{
						A:   nil,
						B:   PackThunk{Path: "foo"},
						Dir: "foo",
					},
				},
			},

			//
		}

		var dump Sink
		dump = func(v any) (Sink, error) {
			if v == nil {
				return nil, nil
			}
			item, ok := v.(ZipItem)
			if !ok {
				panic(fmt.Errorf("expecting ZipItem, got %#v", v))
			}
			pt("==> Path: %s, A %T: %+v, B %T: %+v\n",
				item.Dir,
				item.A,
				item.A,
				item.B,
				item.B,
			)
			return dump, nil
		}
		_ = dump

		check := func(expected []ZipItem) Sink {
			var sink Sink
			sink = func(v any) (Sink, error) {
				if v == nil {
					if len(expected) > 0 {
						return nil, fmt.Errorf("expected %+v, got nil", expected[0])
					} else {
						return nil, nil
					}
				}
				if len(expected) == 0 {
					return nil, fmt.Errorf("unexpected %#v", v)
				}
				e := expected[0]
				i := v.(ZipItem)
				if filepath.ToSlash(e.Dir) != filepath.ToSlash(i.Dir) {
					return nil, fmt.Errorf("expected path %s, got %s", e.Dir, i.Dir)
				}
				if e.A != nil && i.A == nil {
					return nil, fmt.Errorf("expected A %v, got nil", e.A)
				} else if e.A == nil && i.A != nil {
					return nil, fmt.Errorf("expected nil, got %v", i.A)
				} else if e.A != nil && i.A != nil {
					if reflect.TypeOf(e.A) != reflect.TypeOf(i.A) {
						return nil, fmt.Errorf("expected %T, got %T", e.A, i.A)
					}
					switch ex := e.A.(type) {
					case FileInfo:
						vv := i.A.(FileInfo)
						if filepath.ToSlash(ex.Path) != filepath.ToSlash(vv.Path) {
							return nil, fmt.Errorf("expected path %v, got %v", ex.Path, vv.Path)
						}
						if ex.GetName(scope) != vv.GetName(scope) {
							return nil, fmt.Errorf("expected name %v, got %v", ex.GetName(scope), vv.GetName(scope))
						}
					case FileInfoThunk:
						vv := i.A.(FileInfoThunk)
						if filepath.ToSlash(ex.FileInfo.Path) != filepath.ToSlash(vv.FileInfo.Path) {
							return nil, fmt.Errorf("expected path %v, got %v", ex.FileInfo.Path, vv.FileInfo.Path)
						}
						if ex.FileInfo.GetName(scope) != vv.FileInfo.GetName(scope) {
							return nil, fmt.Errorf("expected name %v, got %v", ex.FileInfo.GetName(scope), vv.FileInfo.GetName(scope))
						}
					case PackThunk:
						vv := i.A.(PackThunk)
						if filepath.ToSlash(ex.Path) != filepath.ToSlash(vv.Path) {
							return nil, fmt.Errorf("expected path %v, got %v", ex.Path, vv.Path)
						}
					}
				}
				if e.B != nil && i.B == nil {
					return nil, fmt.Errorf("expected B %v, got nil", e.B)
				} else if e.B == nil && i.B != nil {
					return nil, fmt.Errorf("expected nil, got %v", i.B)
				} else if e.B != nil && i.B != nil {
					if reflect.TypeOf(e.B) != reflect.TypeOf(i.B) {
						return nil, fmt.Errorf("expected %T, got %T", e.B, i.B)
					}
					switch ex := e.B.(type) {
					case FileInfo:
						vv := i.B.(FileInfo)
						if filepath.ToSlash(ex.Path) != filepath.ToSlash(vv.Path) {
							return nil, fmt.Errorf("expected path %v, got %v", ex.Path, vv.Path)
						}
						if ex.GetName(scope) != vv.GetName(scope) {
							return nil, fmt.Errorf("expected name %v, got %v", ex.GetName(scope), vv.GetName(scope))
						}
					case FileInfoThunk:
						vv := i.B.(FileInfoThunk)
						if filepath.ToSlash(ex.FileInfo.Path) != filepath.ToSlash(vv.FileInfo.Path) {
							return nil, fmt.Errorf("expected path %v, got %v", ex.FileInfo.Path, vv.FileInfo.Path)
						}
						if ex.FileInfo.GetName(scope) != vv.FileInfo.GetName(scope) {
							return nil, fmt.Errorf("expected name %v, got %v", ex.FileInfo.GetName(scope), vv.FileInfo.GetName(scope))
						}
					case PackThunk:
						vv := i.B.(PackThunk)
						if filepath.ToSlash(ex.Path) != filepath.ToSlash(vv.Path) {
							return nil, fmt.Errorf("expected path %v, got %v", ex.Path, vv.Path)
						}
					}
				}
				expected = expected[1:]
				return sink, nil
			}
			return sink
		}
		_ = check

		for i, c := range cases {
			if c.Skip {
				continue
			}

			var a, b Src
			switch v := c.A.(type) {
			case Virtual:
				a = iterVirtual(v, nil)
			case Src:
				a = v
			}
			switch v := c.B.(type) {
			case Virtual:
				b = iterVirtual(v, nil)
			case Src:
				b = v
			}

			var items pp.Values
			var valuesA pp.Values
			var valuesB pp.Values
			err := pp.Copy(
				zip(
					pp.Tee(
						a,
						pp.CollectValues(&valuesA),
					),
					pp.Tee(
						b,
						pp.CollectValues(&valuesB),
					),
					nil,
					predictExpand,
				),
				func(v any) (Sink, error) {
					// comment out to show dumps
					return nil, nil
					//pt("---------- %d -----------\n", i)
					//return dump(v)
				},
				check(c.Expected),
				pp.CollectValues(&items),
			)
			ce(err, e5.WrapFunc(func(err error) error {
				pt("--- %d: A ---\n", i)
				for _, v := range valuesA {
					pt("%#v\n", v)
				}
				pt("--- B ---\n")
				for _, v := range valuesB {
					pt("%#v\n", v)
				}
				return err
			}))

			res, err := equal(
				valuesA.Iter(nil),
				unzip(items.Iter(nil), func(item ZipItem) any {
					return item.A
				}, nil),
				nil,
			)
			ce(err)
			if !res {
				t.Fatal()
			}

			res, err = equal(
				valuesB.Iter(nil),
				unzip(reverse(items.Iter(nil), nil), func(item ZipItem) any {
					return item.A
				}, nil),
				nil,
			)
			ce(err)
			if !res {
				t.Fatal()
			}

		}

	})
}

func TestZipFile(
	t *testing.T,
	zip Zip,
	iterDisk IterDiskFile,
	unzip Unzip,
	equal Equal,
	build Build,
	iterFile IterFile,
	shuffleDir fsys.ShuffleDir,
	wg *pr2.WaitGroup,
) {
	defer he(nil, e5.TestingFatal(t))

	dir := t.TempDir()
	for i := 0; i < 64; i++ {
		_, _, _, err := shuffleDir(dir)
		ce(err)

		numFiles := 0
		ce(filepath.WalkDir(dir, func(_ string, _ fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			numFiles++
			return nil
		}))

		// zip(disk ,disk) : disk
		zipped := zip(
			iterDisk(dir, nil),
			iterDisk(dir, nil),
			nil,
		)
		left := unzip(
			zipped,
			func(item ZipItem) any {
				return item.A
			},
			nil,
		)
		ok, err := equal(
			left,
			iterDisk(dir, nil),
			func(a, b any, reason string) {
				pt("DIFF %s\n\t%#v\n\t%#v\n\n", reason, a, b)
			},
		)
		ce(err)
		if !ok {
			t.Fatal()
		}

		// zip(build(disk), disk) : disk
		file := new(File)
		err = Copy(
			iterDisk(dir, nil),
			build(wg, file, nil),
		)
		ce(err)
		file = file.Subs[0].File
		zipped = zip(
			iterFile(file, nil),
			iterDisk(dir, nil),
			nil,
		)
		left = unzip(
			zipped,
			func(item ZipItem) any {
				return item.A
			},
			nil,
		)
		ok, err = equal(
			left,
			iterDisk(dir, nil),
			func(a, b any, reason string) {
				pt("DIFF %s\n\t%#v\n\t%#v\n\n", reason, a, b)
			},
		)
		ce(err)
		if !ok {
			t.Fatal()
		}

		// tap(zip(disk, disk))
		zipped = zip(
			iterDisk(dir, nil),
			iterDisk(dir, nil),
			nil,
		)
		left = unzip(
			zipped,
			func(item ZipItem) any {
				return item.A
			},
			nil,
		)
		n := 0
		err = Copy(
			left,
			pp.Tap(func(v any) (err error) {
				defer he(&err)
				n++
				if info, ok := v.(FileInfo); ok {
					p := filepath.Join(filepath.Dir(dir), info.Path)
					_, err := os.Stat(p)
					ce(err)
				}
				return nil
			}),
		)
		ce(err)
		if n != numFiles {
			// files in testdata/zip
			t.Fatalf("expecting %d, got %d", numFiles, n)
		}

		// collect(zip(disk, disk))
		zipped = zip(
			iterDisk(dir, nil),
			iterDisk(dir, nil),
			nil,
		)
		left = unzip(
			zipped,
			func(item ZipItem) any {
				return item.A
			},
			nil,
		)
		var values pp.Values
		err = Copy(
			left,
			pp.CollectValues(&values),
		)
		ce(err)
		if len(values) != numFiles {
			t.Fatal()
		}

		// build(collect(zip(disk, disk)))
		root := new(File)
		err = Copy(
			values.Iter(nil),
			build(wg, root, nil),
		)
		ce(err)
		root = root.Subs[0].File
		ok, err = equal(
			iterFile(root, nil),
			iterDisk(dir, nil),
			func(a, b any, reason string) {
				pt("DIFF %s\n\t%#v\n\t%#v\n\n", reason, a, b)
			},
		)
		ce(err)
		if !ok {
			t.Fatal()
		}

		// build(collect(zip(disk, disk))) : disk
		file = new(File)
		err = Copy(
			values.Iter(nil),
			build(wg, file, nil),
		)
		ce(err)
		file = file.Subs[0].File
		ok, err = equal(
			iterFile(file, nil),
			iterDisk(dir, nil),
			func(a, b any, reason string) {
				pt("DIFF %s\n\t%#v\n\t%#v\n\n", reason, a, b)
			},
		)
		ce(err)
		if !ok {
			t.Fatal()
		}

		// build(zip(disk, disk) : disk
		zipped = zip(
			iterDisk(dir, nil),
			iterDisk(dir, nil),
			nil,
		)
		left = unzip(
			zipped,
			func(item ZipItem) any {
				return item.A
			},
			nil,
		)
		file = new(File)
		err = Copy(
			left,
			build(wg, file, nil),
		)
		ce(err)
		file = file.Subs[0].File
		ok, err = equal(
			iterFile(file, nil),
			iterDisk(dir, nil),
			func(a, b any, reason string) {
				pt("DIFF %s\n\t%#v\n\t%#v\n\n", reason, a, b)
			},
		)
		ce(err)
		if !ok {
			t.Fatal()
		}

	}

}

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

	"github.com/reusee/e4"
	"github.com/reusee/june/fsys"
	"github.com/reusee/pp"
)

func TestZip(
	t *testing.T,
	scope Scope,
) {
	defer he(nil, e4.TestingFatal(t))

	scope.Sub(
		func() PackThreshold {
			return 2
		},
	).Call(func(
		zip Zip,
		iterVirtual IterVirtual,
		build Build,
		iterKey IterKey,
		unzip Unzip,
		equal Equal,
		reverse Reverse,
		iterFile IterFile,
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
		VirtualDirItem := func(name string, subs ...Virtual) *IterItem {
			return &IterItem{
				Virtual: &Virtual{
					IsDir: true,
					Name:  name,
					Subs:  subs,
				},
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
		VirtualItem := func(name string) *IterItem {
			return &IterItem{
				Virtual: &Virtual{
					Name: name,
				},
			}
		}
		FileItem := func(path string, v Virtual) *IterItem {
			return &IterItem{
				FileInfo: &FileInfo{
					Path:     path,
					FileLike: v,
				},
			}
		}
		ThunkItem := func(path string, v Virtual) *IterItem {
			return &IterItem{
				FileInfoThunk: &FileInfoThunk{
					Path: path,
					FileInfo: FileInfo{
						Path:     path,
						FileLike: v,
					},
					Expand: func(bool) {
						// ignore
					},
				},
			}
		}

		file1 := new(File)
		err := pp.Copy(
			iterVirtual(D("foo", D("bar", D("baz", D("qux", F("quux"))))), nil),
			build(file1, nil),
		)
		ce(err)
		file1 = file1.Subs[0].File
		file2 := new(File)
		err = pp.Copy(
			iterVirtual(D("foo", D("1", D("baz", D("qux", F("quux"))))), nil),
			build(file2, nil),
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
				A: VirtualItem("foo"),
				B: VirtualItem("foo"),
				Expected: []ZipItem{
					ZipItem{
						A:   FileItem("foo", F("foo")),
						B:   FileItem("foo", F("foo")),
						Dir: ".",
					},
				},
			},

			1: {
				A: VirtualItem("bar"),
				B: VirtualItem("foo"),
				Expected: []ZipItem{
					ZipItem{
						A:   FileItem("bar", F("bar")),
						B:   nil,
						Dir: ".",
					},
					ZipItem{
						A:   nil,
						B:   FileItem("foo", F("foo")),
						Dir: ".",
					},
				},
			},

			2: {
				A: VirtualItem("foo"),
				B: VirtualItem("bar"),
				Expected: []ZipItem{
					ZipItem{
						A:   nil,
						B:   FileItem("bar", F("bar")),
						Dir: ".",
					},
					ZipItem{
						A:   FileItem("foo", F("foo")),
						B:   nil,
						Dir: ".",
					},
				},
			},

			// TreeInfo
			3: {
				A: VirtualDirItem("foo",
					F("foo"),
				),
				B: VirtualDirItem("foo",
					F("foo"),
				),
				Expected: []ZipItem{
					ZipItem{
						A:   ThunkItem("foo", D("foo")),
						B:   ThunkItem("foo", D("foo")),
						Dir: ".",
					},
				},
			},

			4: {
				A: VirtualDirItem("bar",
					F("bar"),
				),
				B: VirtualDirItem("foo",
					F("foo"),
				),
				Expected: []ZipItem{
					ZipItem{
						A:   ThunkItem("bar", D("bar")),
						B:   nil,
						Dir: ".",
					},
					ZipItem{
						A:   nil,
						B:   ThunkItem("foo", D("foo")),
						Dir: ".",
					},
				},
			},

			5: {
				A: VirtualDirItem("foo",
					F("foo"),
				),
				B: VirtualDirItem("bar",
					F("bar"),
				),
				Expected: []ZipItem{
					ZipItem{
						A:   nil,
						B:   ThunkItem("bar", D("bar")),
						Dir: ".",
					},
					ZipItem{
						A:   ThunkItem("foo", D("foo")),
						B:   nil,
						Dir: ".",
					},
				},
			},

			// TreeInfo and FileInfo
			6: {
				A: VirtualDirItem("foo", F("foo")),
				B: ExpandAll(iterVirtual(D("foo", F("foo")), nil)),
				Expected: []ZipItem{
					ZipItem{
						A:   FileItem("foo", D("foo")),
						B:   FileItem("foo", D("foo")),
						Dir: ".",
					},
					ZipItem{
						A:   FileItem("foo/foo", D("foo")),
						B:   FileItem("foo/foo", D("foo")),
						Dir: "foo",
					},
				},
			},

			7: {
				A: D("foo", F("bar")),
				B: ExpandAll(iterVirtual(D("foo", F("foo")), nil)),
				Expected: []ZipItem{
					ZipItem{
						A:   FileItem("foo", D("foo")),
						B:   FileItem("foo", D("foo")),
						Dir: ".",
					},
					ZipItem{
						A:   FileItem("foo/bar", D("bar")),
						B:   nil,
						Dir: "foo",
					},
					ZipItem{
						A:   nil,
						B:   FileItem("foo/foo", D("foo")),
						Dir: "foo",
					},
				},
			},

			8: {
				A: ExpandAll(iterVirtual(D("foo", F("foo")), nil)),
				B: D("foo", F("bar")),
				Expected: []ZipItem{
					ZipItem{
						A:   FileItem("foo", D("foo")),
						B:   FileItem("foo", D("foo")),
						Dir: ".",
					},
					ZipItem{
						A:   nil,
						B:   FileItem("foo/bar", D("bar")),
						Dir: "foo",
					},
					ZipItem{
						A:   FileItem("foo/foo", D("foo")),
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
					ZipItem{
						A:   FileItem("foo", D("foo")),
						B:   FileItem("foo", D("foo")),
						Dir: ".",
					},
					ZipItem{
						A:   FileItem("foo/foo", F("foo")),
						B:   FileItem("foo/foo", F("foo")),
						Dir: "foo",
					},
				},
			},

			// dir and file
			10: {
				A: D("foo"),
				B: F("foo"),
				Expected: []ZipItem{
					ZipItem{
						A:   FileItem("foo", D("foo")),
						B:   FileItem("foo", F("foo")),
						Dir: ".",
					},
				},
			},

			11: {
				A: F("foo"),
				B: D("foo"),
				Expected: []ZipItem{
					ZipItem{
						A:   FileItem("foo", F("foo")),
						B:   FileItem("foo", D("foo")),
						Dir: ".",
					},
				},
			},

			12: {
				A: F("foo"),
				B: D("foo", F("foo")),
				Expected: []ZipItem{
					ZipItem{
						A:   FileItem("foo", F("foo")),
						B:   FileItem("foo", D("foo")),
						Dir: ".",
					},
					ZipItem{
						A:   nil,
						B:   FileItem("foo/foo", F("foo")),
						Dir: "foo",
					},
				},
			},

			13: {
				A: D("foo", F("foo")),
				B: F("foo"),
				Expected: []ZipItem{
					ZipItem{
						A:   FileItem("foo", F("foo")),
						B:   FileItem("foo", D("foo")),
						Dir: ".",
					},
					ZipItem{
						A:   FileItem("foo/foo", F("foo")),
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
					ZipItem{
						A:   FileItem("foo", D("foo")),
						B:   FileItem("foo", D("foo")),
						Dir: ".",
					},
					ZipItem{
						A:   FileItem("foo/bar", D("bar")),
						B:   nil,
						Dir: "foo",
					},
					ZipItem{
						A:   FileItem("foo/bar/baz", D("baz")),
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
					ZipItem{
						A:   FileItem("foo", D("foo")),
						B:   FileItem("foo", D("foo")),
						Dir: ".",
					},
					ZipItem{
						A:   nil,
						B:   FileItem("foo/bar", D("bar")),
						Dir: "foo",
					},
					ZipItem{
						A:   nil,
						B:   FileItem("foo/bar/baz", F("baz")),
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
					ZipItem{
						A:   FileItem("bar", D("bar")),
						B:   nil,
						Dir: ".",
					},
					ZipItem{
						A:   nil,
						B:   FileItem("foo", D("foo")),
						Dir: ".",
					},
					ZipItem{
						A:   nil,
						B:   FileItem("foo/bar", D("bar")),
						Dir: "foo",
					},
					ZipItem{
						A:   nil,
						B:   FileItem("foo/bar/baz", F("baz")),
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
					ZipItem{
						A:   nil,
						B:   FileItem("bar", D("bar")),
						Dir: ".",
					},
					ZipItem{
						A:   FileItem("foo", D("foo")),
						B:   nil,
						Dir: ".",
					},
					ZipItem{
						A:   FileItem("foo/bar", D("bar")),
						B:   nil,
						Dir: "foo",
					},
					ZipItem{
						A:   FileItem("foo/bar/baz", F("baz")),
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
					ZipItem{
						A:   FileItem("1", D("1")),
						B:   nil,
						Dir: ".",
					},
					ZipItem{
						A:   FileItem("1/11", D("11")),
						B:   nil,
						Dir: "1",
					},
					ZipItem{
						A:   FileItem("1/11/111", F("111")),
						B:   nil,
						Dir: "1/11",
					},
					ZipItem{
						A:   nil,
						B:   FileItem("2", D("2")),
						Dir: ".",
					},
					ZipItem{
						A:   nil,
						B:   FileItem("2/22", D("22")),
						Dir: "2",
					},
					ZipItem{
						A:   nil,
						B:   FileItem("2/22/222", F("222")),
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
					ZipItem{
						A:   nil,
						B:   FileItem("1", D("1")),
						Dir: ".",
					},
					ZipItem{
						A:   nil,
						B:   FileItem("1/11", D("11")),
						Dir: "1",
					},
					ZipItem{
						A:   nil,
						B:   FileItem("1/11/111", F("111")),
						Dir: "1/11",
					},
					ZipItem{
						A:   FileItem("2", D("2")),
						B:   nil,
						Dir: ".",
					},
					ZipItem{
						A:   FileItem("2/22", D("22")),
						B:   nil,
						Dir: "2",
					},
					ZipItem{
						A:   FileItem("2/22/222", F("222")),
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
					ZipItem{
						A:   nil,
						B:   FileItem("1", D("1")),
						Dir: ".",
					},
					ZipItem{
						A:   nil,
						B:   FileItem("1/bar", D("bar")),
						Dir: "1",
					},
					ZipItem{
						A:   nil,
						B:   FileItem("1/bar/baz", F("baz")),
						Dir: "1/bar",
					},
					ZipItem{
						A:   FileItem("bar", D("bar")),
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
					ZipItem{
						A:   FileItem("1", D("1")),
						B:   nil,
						Dir: ".",
					},
					ZipItem{
						A:   FileItem("1/bar", D("bar")),
						B:   nil,
						Dir: "1",
					},
					ZipItem{
						A:   FileItem("1/bar/baz", F("baz")),
						B:   nil,
						Dir: "1/bar",
					},
					ZipItem{
						A:   nil,
						B:   FileItem("bar", D("bar")),
						Dir: ".",
					},
				},
			},

			22: {
				A: Seq(
					*FileItem("foo", D("foo")),
					*ThunkItem("foo/bar", D("bar")),
				),
				B: Seq(
					*FileItem("foo", D("foo")),
				),
				Expected: []ZipItem{
					ZipItem{
						A:   FileItem("foo", D("foo")),
						B:   FileItem("foo", D("foo")),
						Dir: ".",
					},
					ZipItem{
						A:   ThunkItem("foo/bar", D("bar")),
						B:   nil,
						Dir: "foo",
					},
				},
			},

			23: {
				A: Seq(
					*FileItem("foo", D("foo")),
				),
				B: Seq(
					*FileItem("foo", D("foo")),
					*ThunkItem("foo/bar", D("bar")),
				),
				Expected: []ZipItem{
					ZipItem{
						A:   FileItem("foo", D("foo")),
						B:   FileItem("foo", D("foo")),
						Dir: ".",
					},
					ZipItem{
						A:   nil,
						B:   ThunkItem("foo/bar", D("bar")),
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
						A:   FileItem("foo", D("foo")),
						B:   FileItem("foo", D("foo")),
						Dir: ".",
					},
					{
						A: &IterItem{
							PackThunk: &PackThunk{
								Path: "foo",
							},
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
						A:   FileItem("foo", D("foo")),
						B:   FileItem("foo", D("foo")),
						Dir: ".",
					},
					{
						A: nil,
						B: &IterItem{
							PackThunk: &PackThunk{
								Path: "foo",
							},
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
						A:   FileItem("foo", D("foo")),
						B:   FileItem("foo", D("foo")),
						Dir: ".",
					},
					{
						A:   FileItem("foo/bar", D("bar")),
						B:   FileItem("foo/bar", D("bar")),
						Dir: "foo",
					},
					{
						A: nil,
						B: &IterItem{
							PackThunk: &PackThunk{Path: "foo/bar"},
						},
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
						A:   FileItem("foo", D("foo")),
						B:   FileItem("foo", D("foo")),
						Dir: ".",
					},
					{
						A:   FileItem("foo/bar", D("bar")),
						B:   FileItem("foo/bar", D("bar")),
						Dir: "foo",
					},
					{
						A: &IterItem{
							PackThunk: &PackThunk{Path: "foo/bar"},
						},
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
						A:   FileItem("foo", D("foo")),
						B:   FileItem("foo", D("foo")),
						Dir: ".",
					},
					{
						A:   nil,
						B:   FileItem("foo/1", D("1")),
						Dir: "foo",
					},
					{
						A: &IterItem{
							PackThunk: &PackThunk{Path: "foo"},
						},
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
						A:   FileItem("foo", D("foo")),
						B:   FileItem("foo", D("foo")),
						Dir: ".",
					},
					{
						A:   FileItem("foo/1", D("1")),
						B:   nil,
						Dir: "foo",
					},
					{
						A: nil,
						B: &IterItem{
							PackThunk: &PackThunk{Path: "foo"},
						},
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
						A:   FileItem("foo", D("foo")),
						B:   FileItem("foo", D("foo")),
						Dir: ".",
					},
					{
						A: nil,
						B: &IterItem{
							PackThunk: &PackThunk{Path: "foo"},
						},
						Dir: "foo",
					},
					{
						A:   FileItem("foo/z", D("z")),
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
						A:   FileItem("foo", D("foo")),
						B:   FileItem("foo", D("foo")),
						Dir: ".",
					},
					{
						A: &IterItem{
							PackThunk: &PackThunk{Path: "foo"},
						},
						B:   nil,
						Dir: "foo",
					},
					{
						A:   nil,
						B:   FileItem("foo/z", D("z")),
						Dir: "foo",
					},
				},
			},

			32: {
				A: iterFile(file1, nil),
				B: Seq(
					*FileItem("foo", D("foo")),
					*ThunkItem("foo/bar", D("bar")),
				),
				Expected: []ZipItem{
					{
						A:   FileItem("foo", D("foo")),
						B:   FileItem("foo", D("foo")),
						Dir: ".",
					},
					{
						A:   ThunkItem("foo/bar", D("bar")),
						B:   ThunkItem("foo/bar", D("bar")),
						Dir: "foo",
					},
				},
			},

			33: {
				A: Seq(
					*FileItem("foo", D("foo")),
					*ThunkItem("foo/bar", D("bar")),
				),
				B: iterFile(file1, nil),
				Expected: []ZipItem{
					{
						A:   FileItem("foo", D("foo")),
						B:   FileItem("foo", D("foo")),
						Dir: ".",
					},
					{
						A:   ThunkItem("foo/bar", D("bar")),
						B:   ThunkItem("foo/bar", D("bar")),
						Dir: "foo",
					},
				},
			},

			34: {
				A: Seq(
					*FileItem("foo", D("foo")),
					*ThunkItem("foo/1", D("1")),
				),
				B: iterFile(file1, nil),
				Expected: []ZipItem{
					{
						A:   FileItem("foo", D("foo")),
						B:   FileItem("foo", D("foo")),
						Dir: ".",
					},
					{
						A:   ThunkItem("foo/1", D("1")),
						B:   nil,
						Dir: "foo",
					},
					{
						A: nil,
						B: &IterItem{
							PackThunk: &PackThunk{Path: "foo"},
						},
						Dir: "foo",
					},
				},
			},

			35: {
				A: iterFile(file1, nil),
				B: Seq(
					*FileItem("foo", D("foo")),
					*ThunkItem("foo/1", D("1")),
				),
				Expected: []ZipItem{
					{
						A:   FileItem("foo", D("foo")),
						B:   FileItem("foo", D("foo")),
						Dir: ".",
					},
					{
						A:   nil,
						B:   ThunkItem("foo/1", D("1")),
						Dir: "foo",
					},
					{
						A: &IterItem{
							PackThunk: &PackThunk{Path: "foo"},
						},
						B:   nil,
						Dir: "foo",
					},
				},
			},

			36: {
				A: iterFile(file1, nil),
				B: Seq(
					*FileItem("foo", D("foo")),
					*ThunkItem("foo/z", D("z")),
				),
				Expected: []ZipItem{
					{
						A:   FileItem("foo", D("foo")),
						B:   FileItem("foo", D("foo")),
						Dir: ".",
					},
					{
						A: &IterItem{
							PackThunk: &PackThunk{Path: "foo"},
						},
						B:   nil,
						Dir: "foo",
					},
					{
						A:   nil,
						B:   ThunkItem("foo/z", D("z")),
						Dir: "foo",
					},
				},
			},

			37: {
				A: Seq(
					*FileItem("foo", D("foo")),
					*ThunkItem("foo/z", D("z")),
				),
				B: iterFile(file1, nil),
				Expected: []ZipItem{
					{
						A:   FileItem("foo", D("foo")),
						B:   FileItem("foo", D("foo")),
						Dir: ".",
					},
					{
						A: nil,
						B: &IterItem{
							PackThunk: &PackThunk{Path: "foo"},
						},
						Dir: "foo",
					},
					{
						A:   ThunkItem("foo/z", D("z")),
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
						A:   FileItem("foo", D("foo")),
						B:   FileItem("foo", D("foo")),
						Dir: ".",
					},
					{
						A:   FileItem("foo/bar", D("bar")),
						B:   FileItem("foo/bar", D("bar")),
						Dir: "foo",
					},
					{
						A:   FileItem("foo/bar/baz", D("baz")),
						B:   FileItem("foo/bar/baz", D("baz")),
						Dir: "foo/bar",
					},
					{
						A:   FileItem("foo/bar/baz/qux", D("qux")),
						B:   FileItem("foo/bar/baz/qux", D("qux")),
						Dir: "foo/bar/baz",
					},
					{
						A:   FileItem("foo/bar/baz/qux/quux", F("quux")),
						B:   FileItem("foo/bar/baz/qux/quux", F("quux")),
						Dir: "foo/bar/baz/qux",
					},
				},
			},

			39: {
				A: ExpandFileInfoThunk(iterFile(file1, nil)),
				B: ExpandFileInfoThunk(iterFile(file2, nil)),
				Expected: []ZipItem{
					{
						A:   FileItem("foo", D("foo")),
						B:   FileItem("foo", D("foo")),
						Dir: ".",
					},
					{
						A: nil,
						B: &IterItem{
							PackThunk: &PackThunk{Path: "foo"},
						},
						Dir: "foo",
					},
					{
						A: &IterItem{
							PackThunk: &PackThunk{Path: "foo"},
						},
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
						A:   FileItem("foo", D("foo")),
						B:   FileItem("foo", D("foo")),
						Dir: ".",
					},
					{
						A: &IterItem{
							PackThunk: &PackThunk{Path: "foo"},
						},
						B:   nil,
						Dir: "foo",
					},
					{
						A: nil,
						B: &IterItem{
							PackThunk: &PackThunk{Path: "foo"},
						},
						Dir: "foo",
					},
				},
			},

			//
		}

		var dump Sink
		dump = func(v *IterItem) (Sink, error) {
			if v == nil {
				return nil, nil
			}
			if v.ZipItem == nil {
				panic(fmt.Errorf("expecting ZipItem"))
			}
			item := *v.ZipItem
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
			sink = func(v *IterItem) (Sink, error) {
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
				i := *v.ZipItem
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
					if e.A.FileInfo != nil {
						vv := *i.A.FileInfo
						if filepath.ToSlash(e.A.FileInfo.Path) != filepath.ToSlash(vv.Path) {
							return nil, fmt.Errorf("expected path %v, got %v", e.A.FileInfo.Path, vv.Path)
						}
						if e.A.FileInfo.GetName(scope) != vv.GetName(scope) {
							return nil, fmt.Errorf("expected name %v, got %v", e.A.FileInfo.GetName(scope), vv.GetName(scope))
						}
					} else if e.A.FileInfoThunk != nil {
						vv := *i.A.FileInfoThunk
						if filepath.ToSlash(e.A.FileInfoThunk.FileInfo.Path) != filepath.ToSlash(vv.FileInfo.Path) {
							return nil, fmt.Errorf("expected path %v, got %v", e.A.FileInfoThunk.FileInfo.Path, vv.FileInfo.Path)
						}
						if e.A.FileInfoThunk.FileInfo.GetName(scope) != vv.FileInfo.GetName(scope) {
							return nil, fmt.Errorf("expected name %v, got %v", e.A.FileInfoThunk.FileInfo.GetName(scope), vv.FileInfo.GetName(scope))
						}
					} else if e.A.PackThunk != nil {
						vv := *i.A.PackThunk
						if filepath.ToSlash(e.A.PackThunk.Path) != filepath.ToSlash(vv.Path) {
							return nil, fmt.Errorf("expected path %v, got %v", e.A.PackThunk.Path, vv.Path)
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
					if e.B.FileInfo != nil {
						vv := *i.B.FileInfo
						if filepath.ToSlash(e.B.FileInfo.Path) != filepath.ToSlash(vv.Path) {
							return nil, fmt.Errorf("expected path %v, got %v", e.B.FileInfo.Path, vv.Path)
						}
						if e.B.FileInfo.GetName(scope) != vv.GetName(scope) {
							return nil, fmt.Errorf("expected name %v, got %v", e.B.FileInfo.GetName(scope), vv.GetName(scope))
						}
					} else if e.B.FileInfoThunk != nil {
						vv := *i.B.FileInfoThunk
						if filepath.ToSlash(e.B.FileInfoThunk.FileInfo.Path) != filepath.ToSlash(vv.FileInfo.Path) {
							return nil, fmt.Errorf("expected path %v, got %v", e.B.FileInfoThunk.FileInfo.Path, vv.FileInfo.Path)
						}
						if e.B.FileInfoThunk.FileInfo.GetName(scope) != vv.FileInfo.GetName(scope) {
							return nil, fmt.Errorf("expected name %v, got %v", e.B.FileInfoThunk.FileInfo.GetName(scope), vv.FileInfo.GetName(scope))
						}
					} else if e.B.PackThunk != nil {
						vv := *i.B.PackThunk
						if filepath.ToSlash(e.B.PackThunk.Path) != filepath.ToSlash(vv.Path) {
							return nil, fmt.Errorf("expected path %v, got %v", e.B.PackThunk.Path, vv.Path)
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

			var items Values
			var valuesA Values
			var valuesB Values
			err := Copy(
				zip(
					pp.Tee(
						a,
						CollectValues(&valuesA),
					),
					pp.Tee(
						b,
						CollectValues(&valuesB),
					),
					nil,
					predictExpand,
				),
				func(v *IterItem) (Sink, error) {
					// comment out to show dumps
					return nil, nil
					//pt("---------- %d -----------\n", i)
					//return dump(v)
				},
				check(c.Expected),
				CollectValues(&items),
			)
			ce(err, func(err error) error {
				pt("--- %d: A ---\n", i)
				for _, v := range valuesA {
					pt("%#v\n", v)
				}
				pt("--- B ---\n")
				for _, v := range valuesB {
					pt("%#v\n", v)
				}
				return err
			})

			res, err := equal(
				IterValues(valuesA, nil),
				unzip(IterValues(items, nil), func(item ZipItem) *IterItem {
					return item.A
				}, nil),
				nil,
			)
			ce(err)
			if !res {
				t.Fatal()
			}

			res, err = equal(
				IterValues(valuesB, nil),
				unzip(reverse(IterValues(items, nil), nil), func(item ZipItem) *IterItem {
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
	iterKey IterKey,
	iterFile IterFile,
	shuffleDir fsys.ShuffleDir,
) {
	defer he(nil, e4.TestingFatal(t))

	dir := t.TempDir()
	for i := 0; i < 64; i++ {
		_, _, _, err := shuffleDir(dir)
		ce(err)

		numFiles := 0
		ce(filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
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
			func(item ZipItem) *IterItem {
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
			build(file, nil),
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
			func(item ZipItem) *IterItem {
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
			func(item ZipItem) *IterItem {
				return item.A
			},
			nil,
		)
		n := 0
		err = Copy(
			left,
			TapSink(func(v IterItem) (err error) {
				defer he(&err)
				n++
				if v.FileInfo != nil {
					info := *v.FileInfo
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
			func(item ZipItem) *IterItem {
				return item.A
			},
			nil,
		)
		var values Values
		err = Copy(
			left,
			CollectValues(&values),
		)
		ce(err)
		if len(values) != numFiles {
			t.Fatal()
		}

		// build(collect(zip(disk, disk)))
		root := new(File)
		err = Copy(
			IterValues(values, nil),
			build(root, nil),
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
			IterValues(values, nil),
			build(file, nil),
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
			func(item ZipItem) *IterItem {
				return item.A
			},
			nil,
		)
		file = new(File)
		err = Copy(
			left,
			build(file, nil),
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

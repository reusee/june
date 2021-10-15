// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package file

import (
	"sync/atomic"
	"testing"

	"github.com/reusee/e4"
	"github.com/reusee/june/fsys"
	"github.com/reusee/pp"
)

func TestBuild(
	t *testing.T,
	scope Scope,
	shuffleDir fsys.ShuffleDir,
) {
	defer he(nil, e4.TestingFatal(t))

	scope.Fork(
		func() PackThreshold {
			return 4
		},
	).Call(func(
		build Build,
		iterVirtual IterVirtual,
		iterDisk IterDiskFile,
		iterFile IterFile,
		equal Equal,
	) {

		var root File

		// one file
		ce(pp.Copy(
			iterVirtual(Virtual{
				Name: "foo",
			}, nil),
			build(&root, nil),
		))
		if len(root.Subs) != 1 {
			t.Fatal()
		}
		if root.Subs[0].File == nil {
			t.Fatal()
		}
		if root.Subs[0].File.Name != "foo" {
			t.Fatal()
		}

		// re-use root
		ce(pp.Copy(
			iterVirtual(Virtual{
				Name: "foo",
			}, nil),
			build(&root, nil),
		))
		if len(root.Subs) != 1 {
			t.Fatal()
		}
		if root.Subs[0].File == nil {
			t.Fatal()
		}
		if root.Subs[0].File.Name != "foo" {
			t.Fatal()
		}

		// prepare disk file
		dir := t.TempDir()
		for i := 0; i < 64; i++ {
			_, _, _, err := shuffleDir(dir)
			ce(err)
		}

		// disk file
		var numFile int64
		ce(pp.Copy(
			iterDisk(dir, nil),
			build(
				&root, nil,
				TapBuildFile(func(info FileInfo, file *File) {
					atomic.AddInt64(&numFile, 1)
				}),
			),
		))
		if numFile == 0 {
			t.Fatal()
		}

		// compare
		ok, err := equal(
			iterDisk(dir, nil),
			iterFile(root.Subs[0].File, nil),
			nil,
		)
		ce(err)
		if !ok {
			t.Fatal()
		}

		// build from file
		var root2 File
		ce(pp.Copy(
			iterFile(root.Subs[0].File, nil),
			build(&root2, nil),
		))
		ok, err = equal(
			iterFile(root2.Subs[0].File, nil),
			iterFile(root.Subs[0].File, nil),
			nil,
		)
		ce(err)
		if !ok {
			t.Fatal()
		}

	})

}

func TestBuildMerge(
	t *testing.T,
	scope Scope,
) {

	scope.Fork(
		func() PackThreshold {
			return 2
		},
	).Call(func(
		build Build,
		iterVirtual IterVirtual,
		equal Equal,
		iterFile IterFile,
	) {
		defer he(nil, e4.TestingFatal(t))

		vd := func(name string, subs ...Virtual) Virtual {
			return Virtual{
				Name:  name,
				Subs:  subs,
				IsDir: true,
			}
		}
		vf := func(name string) Virtual {
			return Virtual{
				Name: name,
			}
		}

		vRoot := vd("root",
			vf("f1"),
		)
		var root File

		check := func() {
			ce(pp.Copy(
				iterVirtual(vRoot, nil, NoSubsSort(true)),
				build(&root, nil),
			))
			ok, err := equal(
				iterFile(root.Subs[0].File, nil),
				iterVirtual(vRoot, nil),
				nil,
			)
			ce(err)
			if !ok {
				t.Fatal()
			}
		}
		check()

		vRoot.Subs = append(vRoot.Subs, vf("f2"))
		check()

		vRoot.Subs = append(vRoot.Subs, vf("f3"))
		check()

		vRoot.Subs = append(vRoot.Subs, vf("f5"))
		check()

		vRoot.Subs = append(vRoot.Subs, vf("f4"))
		check()

		vRoot.Subs = append(vRoot.Subs, vf("e1"))
		check()

	})

}

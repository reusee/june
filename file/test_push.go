// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package file

import (
	"testing"

	"github.com/reusee/e4"
	"github.com/reusee/june/entity"
	"github.com/reusee/june/fsys"
	"github.com/reusee/june/index"
	"github.com/reusee/june/storekv"
	"github.com/reusee/june/storemem"
	"github.com/reusee/pp"
	"github.com/reusee/pr"
)

func TestPushFile(
	t *testing.T,
	wt *pr.WaitTree,
	build Build,
	iterDisk IterDiskFile,
	newMem storemem.New,
	push entity.Push,
	save entity.SaveEntity,
	newKV storekv.New,
	indexManager index.IndexManager,
	scope Scope,
	iterFile IterFile,
	shuffleDir fsys.ShuffleDir,
) {
	defer he(nil, e4.TestingFatal(t))

	dir := t.TempDir()
	for i := 0; i < 64; i++ {
		_, _, _, err := shuffleDir(dir)
		ce(err)
	}

	file := new(File)
	ce(pp.Copy(
		iterDisk(dir, nil),
		build(file, nil),
	))
	summary, err := save(file)
	ce(err)

	mem := newMem(wt)
	kv, err := newKV(
		mem, "foo",
	)
	ce(err)
	ce(push(
		kv,
		indexManager,
		[]Key{
			summary.Key,
		},
	))

	scope.Sub(func() Store {
		return kv
	}).Call(func(
		fetch entity.Fetch,
	) {
		file2 := new(File)
		ce(fetch(summary.Key, &file2))
		file3 := new(File)
		ce(pp.Copy(
			iterFile(file2.Subs[0].File, nil),
			build(file3, nil),
		))
		s, err := save(file3)
		ce(err)
		if s.Key != summary.Key {
			t.Fatal()
		}
	})

}

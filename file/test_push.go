// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package file

import (
	"testing"

	"github.com/reusee/e5"
	"github.com/reusee/june/entity"
	"github.com/reusee/june/fsys"
	"github.com/reusee/june/index"
	"github.com/reusee/june/storekv"
	"github.com/reusee/june/storemem"
	"github.com/reusee/pp"
	"github.com/reusee/pr2"
)

func TestPushFile(
	t *testing.T,
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
	wg *pr2.WaitGroup,
) {
	defer he(nil, e5.TestingFatal(t))

	dir := t.TempDir()
	for i := 0; i < 64; i++ {
		_, _, _, err := shuffleDir(dir)
		ce(err)
	}

	file := new(File)
	ce(pp.Copy(
		iterDisk(dir, nil),
		build(wg, file, nil),
	))
	summary, err := save(wg, file)
	ce(err)

	mem := newMem(wg)
	kv, err := newKV(
		mem, "foo",
	)
	ce(err)
	ce(push(
		wg,
		kv,
		indexManager,
		[]Key{
			summary.Key,
		},
	))

	scope.Fork(func() Store {
		return kv
	}).Call(func(
		fetch entity.Fetch,
	) {
		file2 := new(File)
		ce(fetch(summary.Key, &file2))
		file3 := new(File)
		ce(pp.Copy(
			iterFile(file2.Subs[0].File, nil),
			build(wg, file3, nil),
		))
		s, err := save(wg, file3)
		ce(err)
		if s.Key != summary.Key {
			t.Fatal()
		}
	})

}

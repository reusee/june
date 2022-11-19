// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package file

import (
	"context"
	"testing"

	"github.com/reusee/e5"
	"github.com/reusee/june/entity"
	"github.com/reusee/june/fsys"
	"github.com/reusee/june/index"
	"github.com/reusee/june/storekv"
	"github.com/reusee/june/storemem"
	"github.com/reusee/pp"
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
) {
	defer he(nil, e5.TestingFatal(t))
	ctx := context.Background()

	dir := t.TempDir()
	for i := 0; i < 64; i++ {
		_, _, _, err := shuffleDir(dir)
		ce(err)
	}

	file := new(File)
	ce(pp.Copy(
		iterDisk(ctx, dir, nil),
		build(ctx, file, nil),
	))
	summary, err := save(ctx, file)
	ce(err)

	mem := newMem()
	kv, err := newKV(
		mem, "foo",
	)
	ce(err)
	ce(push(
		ctx,
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
		ce(fetch(ctx, summary.Key, &file2))
		file3 := new(File)
		ce(pp.Copy(
			iterFile(ctx, file2.Subs[0].File, nil),
			build(ctx, file3, nil),
		))
		s, err := save(ctx, file3)
		ce(err)
		if s.Key != summary.Key {
			t.Fatal()
		}
	})

}

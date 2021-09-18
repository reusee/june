// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package file

import (
	"context"
	"testing"

	"github.com/reusee/pr"
)

func TestIterIgnore(
	t *testing.T,
	scope Scope,
) {

	scope.Sub(
		func() Ignore {
			return func(path string, file FileLike) bool {
				return file.GetName(scope) == "foo"
			}
		},
	).Call(func(
		iter IterVirtual,
	) {

		var count Sink
		n := 0
		count = func(v *IterItem) (Sink, error) {
			if v == nil {
				return nil, nil
			}
			n++
			if v.FileInfoThunk != nil {
				v.FileInfoThunk.Expand(true)
			}
			return count, nil
		}
		if err := Copy(
			iter(
				Virtual{
					IsDir: true,
					Subs: []Virtual{
						{
							Name: "foo",
						},
						{
							Name: "bar",
						},
					},
				},
				nil,
			),
			count,
		); err != nil {
			t.Fatal(err)
		}
		if n != 2 {
			t.Fatalf("got %d", n)
		}
	})

}

func TestIterDiskCancelWaitTree(
	t *testing.T,
	scope Scope,
	parentWt *pr.WaitTree,
) {
	wt := pr.NewWaitTree(parentWt)
	wt.Cancel()
	scope.Sub(func() *pr.WaitTree {
		return wt
	}).Call(func(
		iterDisk IterDiskFile,
	) {
		err := Copy(
			iterDisk(".", nil),
			Discard,
		)
		if !is(err, context.Canceled) {
			t.Fatal()
		}
	})
}

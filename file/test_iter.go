// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package file

import (
	"context"
	"testing"

	"github.com/reusee/pp"
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
		count = func(v any) (Sink, error) {
			if v == nil {
				return nil, nil
			}
			n++
			if t, ok := v.(FileInfoThunk); ok {
				t.Expand(true)
			}
			return count, nil
		}
		if err := pp.Copy(
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

func TestIterDiskCancelCtx(
	t *testing.T,
	scope Scope,
	ctx context.Context,
) {
	ctx, cancel := context.WithCancel(ctx)
	cancel()
	scope.Sub(func() context.Context {
		return ctx
	}).Call(func(
		iterDisk IterDiskFile,
	) {
		err := pp.Copy(
			iterDisk(".", nil),
			pp.Discard,
		)
		if !is(err, context.Canceled) {
			t.Fatal()
		}
	})

}

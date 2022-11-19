// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storemonotree

import (
	"context"

	"github.com/reusee/june/key"
	"github.com/reusee/june/store"
	"github.com/reusee/sb"
)

var _ store.Store = new(Tree)

func (t *Tree) ID(ctx context.Context) (store.ID, error) {
	return t.id, nil
}

func (t *Tree) Name() string {
	return "monotree(" + t.upstream.Name() + ")"
}

func (t *Tree) Write(
	ctx context.Context,
	ns key.Namespace,
	stream sb.Stream,
	options ...store.WriteOption,
) (
	res store.WriteResult,
	err error,
) {
	defer he(&err)

	//TODO

	return
}

func (t *Tree) Read(
	ctx context.Context,
	key store.Key,
	fn func(sb.Stream) error,
) (
	err error,
) {
	defer he(&err)

	//TODO

	return
}

func (t *Tree) Exists(
	ctx context.Context,
	key store.Key,
) (
	exists bool,
	err error,
) {
	defer he(&err)

	//TODO

	return
}

func (t *Tree) IterKeys(
	ctx context.Context,
	ns key.Namespace,
	fn func(store.Key) error,
) (
	err error,
) {
	defer he(&err)

	//TODO

	return
}

func (t *Tree) IterAllKeys(
	ctx context.Context,
	fn func(store.Key) error,
) (
	err error,
) {
	defer he(&err)

	//TODO

	return
}

func (t *Tree) Delete(
	ctx context.Context,
	keys []store.Key,
) (
	err error,
) {
	defer he(&err)

	//TODO

	return
}

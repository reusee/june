// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storemonotree

import (
	"github.com/reusee/june/key"
	"github.com/reusee/june/store"
	"github.com/reusee/sb"
)

var _ store.Store = new(Tree)

func (t *Tree) ID() (store.ID, error) {
	return t.id, nil
}

func (t *Tree) Name() string {
	return "monotree(" + t.upstream.Name() + ")"
}

func (t *Tree) Write(
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
	fn func(store.Key) error,
) (
	err error,
) {
	defer he(&err)

	//TODO

	return
}

func (t *Tree) Delete(
	keys []store.Key,
) (
	err error,
) {
	defer he(&err)

	//TODO

	return
}

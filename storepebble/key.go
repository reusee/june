// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storepebble

import (
	"bytes"
	"sync"

	"github.com/reusee/sb"
)

type Prefix uint8

const (
	Kv Prefix = iota + 1
	Idx
)

var keyBufPool = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

func withMarshalKey(fn func([]byte), prefix Prefix, args ...any) {
	buf := keyBufPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		keyBufPool.Put(buf)
	}()
	tuple := sb.Tuple{prefix}
	tuple = append(tuple, args...)
	if err := sb.Copy(
		sb.Marshal(tuple),
		sb.Encode(buf),
	); err != nil {
		panic(err)
	}
	fn(buf.Bytes())
}

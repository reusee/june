// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package codec

import "github.com/reusee/sb"

type defaultCodec struct{}

// no-op codec
var DefaultCodec = defaultCodec{}

var _ Codec = defaultCodec{}

func (d defaultCodec) ID() string {
	return ""
}

func (d defaultCodec) Encode(sink sb.Sink, options ...Option) sb.Sink {
	return sink
}

func (d defaultCodec) Decode(proc sb.Proc, options ...Option) sb.Proc {
	return proc
}

// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package codec

import "github.com/reusee/sb"

type stackedCodec struct {
	a Codec
	b Codec
}

func Stacked(a, b Codec) stackedCodec {
	return stackedCodec{
		a: a,
		b: b,
	}
}

var _ Codec = stackedCodec{}

func (s stackedCodec) Encode(sink sb.Sink, options ...Option) sb.Sink {
	return s.a.Encode(s.b.Encode(sink, options...), options...)
}

func (s stackedCodec) Decode(src sb.Proc, options ...Option) sb.Proc {
	return s.a.Decode(s.b.Decode(src, options...), options...)
}

func (s stackedCodec) ID() string {
	return s.a.ID() + "|" + s.b.ID()
}

// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package codec

import (
	"github.com/reusee/sb"
)

type Codec interface {
	Encode(sb.Sink, ...Option) sb.Sink
	Decode(sb.Stream, ...Option) sb.Stream
	ID() string
}

type Option interface {
	IsCodecOption()
}

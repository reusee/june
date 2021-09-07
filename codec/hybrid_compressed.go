// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package codec

import (
	"bytes"
	"fmt"

	"github.com/reusee/june/opts"
	"github.com/reusee/sb"
)

func HybridCompressed(
	compressType string,
	compress CompressFunc,
	uncompress UncompressFunc,
) hybridCompressedCodec {
	return hybridCompressedCodec{
		compress:   compress,
		uncompress: uncompress,
		id:         fmt.Sprintf("hybrid-compressed(%s)", compressType),
	}
}

type hybridCompressedCodec struct {
	compress   CompressFunc
	uncompress UncompressFunc
	id         string
}

var _ Codec = hybridCompressedCodec{}

func (c hybridCompressedCodec) ID() string {
	return c.id
}

func (c hybridCompressedCodec) Encode(sink sb.Sink, options ...Option) sb.Sink {
	var buf, buf2 *bytes.Buffer
	for _, option := range options {
		switch option := option.(type) {
		case opts.NewBytesBuffer:
			buf = option()
			buf2 = option()
		}
	}
	if buf == nil {
		buf = new(bytes.Buffer)
		buf2 = new(bytes.Buffer)
	}
	return sb.EncodeBuffer(
		buf,
		make([]byte, 8),
		func(_ *sb.Token) (sb.Sink, error) {
			src := buf.Bytes()
			if err := c.compress(src, buf2); err != nil {
				return nil, err
			}
			var bs []byte
			compressed := false
			if buf2.Len() < len(src) {
				bs = buf2.Bytes()
				compressed = true
			} else {
				bs = src
			}
			if err := sb.Copy(
				sb.Marshal(func() (bool, []byte) {
					return compressed, bs
				}),
				sink,
			); err != nil {
				return nil, err
			}
			return nil, nil
		},
	)
}

func (c hybridCompressedCodec) Decode(src sb.Proc, options ...Option) sb.Proc {
	proc := sb.Proc(func() (_ *sb.Token, _ sb.Proc, err error) {
		defer he(&err)
		var compressed bool
		var bs []byte
		if err := sb.Copy(
			src,
			sb.Unmarshal(func(c bool, b []byte) {
				compressed = c
				bs = b
			}),
		); err != nil {
			return nil, nil, err
		}
		var data []byte
		if compressed {
			var err error
			data, err = c.uncompress(bs)
			ce(err)
		} else {
			data = bs
		}
		br := bytes.NewReader(data)
		buf := make([]byte, 8)
		return nil, sb.DecodeBuffer(br, br, buf, nil), nil
	})
	return proc
}

func HybridSnappy() hybridCompressedCodec {
	codec := HybridCompressed(UseSnappy())
	codec.id = "hybrid-snappy" // compatible
	return codec
}

func HybridZstd() hybridCompressedCodec {
	codec := HybridCompressed(UseZstd())
	codec.id = "hybrid-zstd" // compatible
	return codec
}

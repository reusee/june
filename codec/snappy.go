// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package codec

import (
	"bytes"

	"github.com/golang/snappy"
	"github.com/reusee/june/opts"
	"github.com/reusee/sb"
)

type snappyCodec struct {
}

var _ Codec = snappyCodec{}

func (a snappyCodec) ID() string {
	return "snappy"
}

func Snappy() snappyCodec {
	return snappyCodec{}
}

func (a snappyCodec) Encode(sink sb.Sink, options ...Option) sb.Sink {
	var buf *bytes.Buffer
	for _, option := range options {
		switch option := option.(type) {
		case opts.NewBytesBuffer:
			buf = option()
		}
	}
	if buf == nil {
		buf = new(bytes.Buffer)
	}
	return sb.EncodeBuffer(
		buf,
		make([]byte, 8),
		func(_ *sb.Token) (sb.Sink, error) {
			src := buf.Bytes()
			dst := make([]byte, snappy.MaxEncodedLen(len(src)))
			dst = snappy.Encode(dst, src)
			if err := sb.Copy(
				sb.Marshal(dst),
				sink,
			); err != nil {
				return nil, err
			}
			return nil, nil
		},
	)
}

func (a snappyCodec) Decode(src sb.Proc, options ...Option) sb.Proc {
	proc := sb.Proc(func() (_ *sb.Token, _ sb.Proc, err error) {
		defer he(&err)
		var compressed []byte
		if err := sb.Copy(
			src,
			sb.Unmarshal(&compressed),
		); err != nil {
			return nil, nil, err
		}
		data, err := snappy.Decode(nil, compressed)
		ce(err)
		br := bytes.NewReader(data)
		buf := make([]byte, 8)
		return nil, sb.DecodeBuffer(br, br, buf, nil), nil
	})
	return proc
}

// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package codec

import (
	"errors"

	"github.com/reusee/e4"
	"github.com/reusee/sb"
)

type OnionCodec struct {
	encoders   []Codec
	encoderIDs []string
	decoders   map[string]Codec
}

var _ Codec = OnionCodec{}

func NewOnionCodec(
	encoders []Codec,
	decoders []Codec,
) OnionCodec {
	var ids []string
	for _, codec := range encoders {
		ids = append(ids, codec.ID())
	}
	m := make(map[string]Codec)
	for _, codec := range decoders {
		id := codec.ID()
		m[id] = codec
	}
	return OnionCodec{
		encoders:   encoders,
		encoderIDs: ids,
		decoders:   m,
	}
}

func (_ OnionCodec) ID() string {
	return "onion"
}

func (o OnionCodec) Encode(sink sb.Sink, options ...Option) sb.Sink {
	headerWritten := false

	var s sb.Sink
	s = func(token *sb.Token) (sb.Sink, error) {
		var err error

		if !headerWritten {
			sink, err = sink(&sb.Token{
				Kind: sb.KindTuple,
			})
			if err != nil {
				return nil, err
			}
			sink, err = sink.Marshal(o.encoderIDs)
			if err != nil {
				return nil, err
			}
			headerWritten = true
		}

		if token == nil {
			sink, err = sink(&sb.Token{
				Kind: sb.KindTupleEnd,
			})
			if err != nil {
				return nil, err
			}
			return sink, nil
		}

		sink, err = sink(token)
		if err != nil {
			return nil, err
		}

		return s, nil
	}

	cur := s
	for i := len(o.encoders) - 1; i >= 0; i-- {
		s := o.encoders[i].Encode(cur, options...)
		cur = s
	}

	return cur
}

var ErrCodecNotFound = errors.New("codec not found")

func (o OnionCodec) Decode(src sb.Proc, options ...Option) sb.Proc {
	proc := sb.Proc(func() (*sb.Token, sb.Proc, error) {
		var ids []string
		var tokens sb.Tokens
		if err := sb.Copy(
			src,
			sb.Unmarshal(&sb.Tuple{
				sb.Unmarshal(&ids),
				sb.CollectValueTokens(&tokens),
			}),
		); err != nil {
			return nil, nil, err
		}
		cur := tokens.Iter()
		for i := len(ids) - 1; i >= 0; i-- {
			id := ids[i]
			codec, ok := o.decoders[id]
			if !ok {
				return nil, nil, we.With(
					e4.NewInfo("no such codec: %s", id),
				)(ErrCodecNotFound)
			}
			s := codec.Decode(cur, options...)
			cur = s
		}
		return nil, sb.Iter(cur, nil), nil
	})
	return proc
}

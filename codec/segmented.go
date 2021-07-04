// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package codec

import (
	"bytes"
	"fmt"
	"io"

	"github.com/reusee/june/opts"
	"github.com/reusee/sb"
)

type segmentedCodec struct {
	wrapped Codec
	id      string
	cutSize int
}

var _ Codec = segmentedCodec{}

func Segmented(
	wrapped Codec,
	cutSize int,
) segmentedCodec {
	return segmentedCodec{
		wrapped: wrapped,
		id:      fmt.Sprintf("segmented(%s)", wrapped.ID()),
		cutSize: cutSize,
	}
}

func (s segmentedCodec) ID() string {
	return s.id
}

func (s segmentedCodec) Encode(sink sb.Sink, options ...Option) sb.Sink {

	var step sb.Sink
	step = func(token *sb.Token) (sb.Sink, error) {
		if token == nil {
			return nil, nil
		}
		var err error
		sink, err = sink(token)
		if err != nil {
			return nil, err
		}
		return step, nil
	}

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

	encodeSink := sb.EncodeBuffer(buf, make([]byte, 8), nil)
	headerWritten := false

	var retSink sb.Sink
	retSink = func(token *sb.Token) (_ sb.Sink, err error) {
		defer he(&err)

		if !headerWritten {
			sink, err = sink(&sb.Token{
				Kind: sb.KindArray,
			})
			ce(err)
			headerWritten = true
		}

		if token == nil {
			if buf.Len() > 0 {
				chunk := make([]byte, buf.Len())
				copy(chunk, buf.Bytes())
				ce(sb.Copy(
					sb.Marshal(chunk),
					s.wrapped.Encode(step, options...),
				))
				buf.Reset()
			}
			sink, err = sink(&sb.Token{
				Kind: sb.KindArrayEnd,
			})
			ce(err)
			return sink, nil
		}

		encodeSink, err = encodeSink(token)
		ce(err)
		if buf.Len() > s.cutSize {
			chunk := make([]byte, buf.Len())
			copy(chunk, buf.Bytes())
			ce(sb.Copy(
				sb.Marshal(chunk),
				s.wrapped.Encode(step, options...),
			))
			buf.Reset()
		}

		return retSink, nil
	}

	return retSink
}

type segmentedReader struct {
	wrapped    Codec
	stream     sb.Stream
	options    []Option
	chunk      []byte
	headerRead bool
}

var _ io.Reader = new(segmentedReader)

func (s *segmentedReader) Read(buf []byte) (int, error) {
	if len(s.chunk) > 0 {
		n := copy(buf, s.chunk)
		s.chunk = s.chunk[n:]
		return n, nil
	}
	if err := s.read(); err != nil {
		return 0, err
	}
	return s.Read(buf)
}

func (s *segmentedReader) read() error {
	token, err := s.stream.Next()
	if err != nil {
		return err
	}
	if token == nil {
		return io.ErrUnexpectedEOF
	}

	if !s.headerRead {
		if token.Kind == sb.KindArray {
			s.headerRead = true
			return nil
		} else {
			return fmt.Errorf("expecting KindArray")
		}
	}

	if token.Kind == sb.KindArrayEnd {
		return io.EOF
	}

	p := sb.Proc(func() (*sb.Token, sb.Proc, error) {
		return token, sb.IterStream(s.stream, nil), nil
	})
	if err := sb.Copy(
		s.wrapped.Decode(&p, s.options...),
		sb.Unmarshal(&s.chunk),
	); err != nil {
		return err
	}

	return nil
}

var _ io.ByteReader = new(segmentedReader)

func (s *segmentedReader) ReadByte() (byte, error) {
	if len(s.chunk) > 0 {
		b := s.chunk[0]
		s.chunk = s.chunk[1:]
		return b, nil
	}
	if err := s.read(); err != nil {
		return 0, err
	}
	return s.ReadByte()
}

func (s segmentedCodec) Decode(stream sb.Stream, options ...Option) sb.Stream {
	r := &segmentedReader{
		wrapped: s.wrapped,
		stream:  stream,
		options: options,
	}
	buf := make([]byte, 8)
	proc := sb.DecodeBuffer(r, r, buf, nil)
	return &proc
}

// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package filebase

import (
	"io"
	"reflect"

	"github.com/reusee/e4"
	"github.com/reusee/fastcdc-go"
	"github.com/reusee/ling/v2/entity"
	"github.com/reusee/sb"
)

type Content []byte

var _ sb.SBMarshaler = Content{}

const contentChunkSize = 4 * 1024

func (c Content) MarshalSB(ctx sb.Ctx, cont sb.Proc) sb.Proc {
	if len(c) <= contentChunkSize {
		return ctx.Marshal(ctx, reflect.ValueOf([]byte(c)), cont)
	}
	start := 0
	var proc sb.Proc
	proc = func() (*sb.Token, sb.Proc, error) {
		end := start + contentChunkSize
		if end >= len(c) {
			end = len(c)
			return nil, ctx.Marshal(
				ctx,
				reflect.ValueOf([]byte(c[start:end])),
				ctx.Marshal(
					ctx,
					reflect.ValueOf(&sb.Token{
						Kind: sb.KindArrayEnd,
					}),
					cont,
				),
			), nil
		}
		i := start
		start += contentChunkSize
		return nil, ctx.Marshal(
			ctx,
			reflect.ValueOf([]byte(c[i:end])),
			proc,
		), nil
	}
	return ctx.Marshal(
		ctx,
		reflect.ValueOf(&sb.Token{
			Kind: sb.KindArray,
		}),
		proc,
	)
}

var _ sb.SBUnmarshaler = new(Content)

func (c *Content) UnmarshalSB(ctx sb.Ctx, cont sb.Sink) sb.Sink {
	return func(token *sb.Token) (sb.Sink, error) {
		if token == nil {
			return nil, we(sb.UnmarshalError,
				e4.With(sb.ExpectingValue),
			)
		}
		if token.Kind == sb.KindBytes {
			*c = Content(token.Value.([]byte))
			return cont, nil
		} else if token.Kind == sb.KindArray {
			var sink sb.Sink
			sink = func(token *sb.Token) (sb.Sink, error) {
				if token == nil {
					return nil, we(sb.UnmarshalError,
						e4.With(sb.ExpectingValue),
					)
				}
				if token.Kind == sb.KindArrayEnd {
					return cont, nil
				} else if token.Kind == sb.KindBytes {
					*c = append(*c, Content(token.Value.([]byte))...)
					return sink, nil
				}
				return nil, we(sb.UnmarshalError,
					e4.With(sb.BadTokenKind),
				)
			}
			return sink, nil
		}
		return nil, we(sb.UnmarshalError,
			e4.With(sb.BadTokenKind),
		)
	}
}

type WriteContents func(
	keys []Key,
	w io.Writer,
) (
	err error,
)

func (_ Def) WriteContents(
	fetch entity.Fetch,
) WriteContents {
	return func(
		keys []Key,
		w io.Writer,
	) (
		err error,
	) {
		defer he(&err)

		for _, key := range keys {
			var content Content
			err = fetch(key, &content)
			ce(err, e4.NewInfo("fetch content %s", key))
			_, err := w.Write(content)
			ce(err)
		}

		return nil
	}
}

type ChunkThreshold int64

type MaxChunkSize int64

func (_ Def) ChunkArgs() (ChunkThreshold, MaxChunkSize) {
	return 1 * 1024 * 1024, 32 * 1024 * 1024
}

type ToContentsWithTx func(fn any)

func (_ Def) ToContentsWithTx() ToContentsWithTx {
	return nil
}

type ToContents func(
	r io.Reader,
	size int64,
) (
	keys []Key,
	lengths []int64,
	err error,
)

func (_ Def) ToContents(
	save entity.SaveEntity,
	chunkThreshold ChunkThreshold,
	maxChunkSize MaxChunkSize,
	withTx ToContentsWithTx,
) ToContents {

	maxSize := int64(maxChunkSize)

	var doSave entity.SaveEntity
	if withTx == nil {
		doSave = save
	} else {
		doSave = func(value any, options ...entity.SaveOption) (
			summary *entity.Summary,
			err error,
		) {
			withTx(func(
				save entity.SaveEntity,
			) {
				summary, err = save(value, options...)
			})
			return
		}
	}

	return func(
		r io.Reader,
		size int64,
	) (
		keys []Key,
		lengths []int64,
		err error,
	) {

		defer he(&err)

		if size < int64(chunkThreshold) {
			// no chunking
			data := make([]byte, size)
			_, err := io.ReadFull(r, data)
			ce(err)
			summary, err := doSave(Content(data))
			ce(err)
			return []Key{summary.Key}, []int64{size}, nil
		}

		max := int(size)
		if max > int(maxSize) {
			max = int(maxSize)
		}
		avg := max / 2
		min := max / 4

		chunker, err := fastcdc.NewChunker(r, fastcdc.Options{
			MinSize:     min,
			AverageSize: avg,
			MaxSize:     max,
			Buffer:      make([]byte, max),
		})
		ce(err)

		for {
			chunk, err := chunker.Next()
			if len(chunk.Data) > 0 {
				summary, err := doSave(Content(chunk.Data))
				ce(err)
				keys = append(keys, summary.Key)
				lengths = append(lengths, int64(len(chunk.Data)))
			}
			if is(err, io.EOF) {
				break
			}
			ce(err)
		}

		return keys, lengths, nil

	}
}

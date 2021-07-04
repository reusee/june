// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package codec

import (
	"bytes"
	"io"
	"runtime"

	"github.com/golang/snappy"
	"github.com/klauspost/compress/zstd"
	"github.com/reusee/pr"
)

type CompressFunc = func([]byte, io.Writer) error

type UncompressFunc = func([]byte) ([]byte, error)

func UseSnappy() (string, CompressFunc, UncompressFunc) {
	return "snappy", func(data []byte, w io.Writer) error {
			dst := make([]byte, snappy.MaxEncodedLen(len(data)))
			data = snappy.Encode(dst, data)
			_, err := w.Write(data)
			return err
		}, func(data []byte) ([]byte, error) {
			data, err := snappy.Decode(nil, data)
			if err != nil {
				return nil, err
			}
			return data, nil
		}
}

func UseSnappyStream() (string, CompressFunc, UncompressFunc) {
	return "snappy-stream", func(data []byte, w io.Writer) error {
			sw := snappy.NewBufferedWriter(w)
			_, err := sw.Write(data)
			if err != nil {
				return err
			}
			if err := sw.Close(); err != nil {
				return err
			}
			return nil
		}, func(data []byte) ([]byte, error) {
			r := snappy.NewReader(bytes.NewReader(data))
			data, err := io.ReadAll(r)
			if err != nil {
				return nil, err
			}
			return data, nil
		}
}

func UseZstd() (string, CompressFunc, UncompressFunc) {
	return "zstd", func(data []byte, w io.Writer) error {
			encV, encPut := zstdEncoderPool.Get()
			defer func() {
				encPut()
			}()
			enc := encV.(*zstd.Encoder)
			enc.Reset(w)
			_, err := enc.Write(data)
			if err != nil {
				return err
			}
			if err := enc.Close(); err != nil {
				return err
			}
			return nil
		}, func(data []byte) ([]byte, error) {
			v, put := zstdDecoderPool.Get()
			r := v.(*zstd.Decoder)
			defer func() {
				if !put() {
					r.Close()
				}
			}()
			if err := r.Reset(bytes.NewReader(data)); err != nil {
				return nil, err
			}
			data, err := io.ReadAll(r)
			if err != nil {
				return nil, err
			}
			return data, nil
		}
}

var zstdEncoderPool = pr.NewPool(
	int32(runtime.NumCPU()),
	func() any {
		enc, err := zstd.NewWriter(nil,
			zstd.WithEncoderCRC(true),
			zstd.WithEncoderLevel(zstd.SpeedDefault),
		)
		ce(err)
		return enc
	},
)

var zstdDecoderPool = pr.NewPool(
	int32(runtime.NumCPU()),
	func() any {
		r, err := zstd.NewReader(nil,
			zstd.WithDecoderMaxMemory(64*1024*1024),
		)
		ce(err)
		return r
	},
)

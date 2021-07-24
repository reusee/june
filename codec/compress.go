// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package codec

import (
	"bytes"
	"io"

	"github.com/golang/snappy"
	"github.com/klauspost/compress/zstd"
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
			enc, err := zstd.NewWriter(
				w,
				zstd.WithEncoderCRC(true),
				zstd.WithEncoderLevel(zstd.SpeedDefault),
			)
			if err != nil {
				return err
			}
			_, err = enc.Write(data)
			if err != nil {
				return err
			}
			if err := enc.Close(); err != nil {
				return err
			}
			return nil
		}, func(data []byte) ([]byte, error) {
			r, err := zstd.NewReader(
				bytes.NewReader(data),
				zstd.WithDecoderMaxMemory(64*1024*1024),
			)
			if err != nil {
				return nil, err
			}
			data, err = io.ReadAll(r)
			if err != nil {
				return nil, err
			}
			return data, nil
		}
}

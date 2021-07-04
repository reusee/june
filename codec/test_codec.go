// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package codec

import (
	"bytes"
	"math/rand"
	"strings"
	"testing"

	"github.com/reusee/sb"
)

func testCodec(
	t *testing.T,
	codec Codec,
) {
	t.Helper()

	for i := 0; i < 1024; i++ {
		buf := new(bytes.Buffer)

		n := rand.Int63()
		if err := sb.Copy(
			sb.Marshal(n),
			codec.Encode(sb.Encode(buf)),
		); err != nil {
			t.Fatal(err)
		}
		var res int64
		if err := sb.Copy(
			codec.Decode(sb.Decode(buf)),
			sb.Unmarshal(&res),
		); err != nil {
			t.Fatal(err)
		}
		if res != n {
			t.Fatal()
		}

		s := strings.Repeat("foo", i)
		buf.Reset()
		if err := sb.Copy(
			sb.Marshal(s),
			codec.Encode(sb.Encode(buf)),
		); err != nil {
			t.Fatal(err)
		}
		var s2 string
		if err := sb.Copy(
			codec.Decode(sb.Decode(buf)),
			sb.Unmarshal(&s2),
		); err != nil {
			t.Fatal(err)
		}
		if s2 != s {
			t.Fatal()
		}

	}
}

func TestCodecAESGCM(
	t *testing.T,
) {
	codec := AESGCM([]byte("1234567890123456"))
	testCodec(t, codec)
}

func TestCodecSnappy(
	t *testing.T,
) {
	codec := Snappy()
	testCodec(t, codec)
}

func TestCodecStacked(
	t *testing.T,
) {
	codec := Stacked(
		Snappy(),
		AESGCM([]byte("1234567890123456")),
	)
	testCodec(t, codec)
	codec = Stacked(
		Snappy(),
		Snappy(),
	)
	testCodec(t, codec)
}

func TestHybridSnappy(
	t *testing.T,
) {
	codec := HybridSnappy()
	testCodec(t, codec)
}

func TestHybridZstd(
	t *testing.T,
) {
	codec := HybridZstd()
	testCodec(t, codec)
}

func TestOnion(
	t *testing.T,
) {
	codec := NewOnionCodec(
		[]Codec{
			Snappy(),
			AESGCM([]byte("1234567890123456")),
		},
		[]Codec{
			Snappy(),
			AESGCM([]byte("1234567890123456")),
		},
	)
	testCodec(t, codec)

	codec = NewOnionCodec(
		[]Codec{
			Snappy(),
			AESGCM([]byte("1234567890123456")),
			Snappy(),
		},
		[]Codec{
			Snappy(),
			AESGCM([]byte("1234567890123456")),
		},
	)
	testCodec(t, codec)
}

func TestOnion2(
	t *testing.T,
) {
	codec := NewOnionCodec(
		[]Codec{
			HybridSnappy(),
			Segmented(
				AESGCM([]byte("1111111111111111")),
				8,
			),
		},
		[]Codec{
			HybridSnappy(),
			Segmented(
				AESGCM([]byte("1111111111111111")),
				8,
			),
		},
	)
	testCodec(t, codec)
}

func TestSegmented(
	t *testing.T,
) {
	codec := Segmented(
		HybridSnappy(),
		8,
	)
	testCodec(t, codec)
	codec = Segmented(
		defaultCodec{},
		8,
	)
	testCodec(t, codec)
	codec = Segmented(
		AESGCM([]byte("1234567890123456")),
		8,
	)
	testCodec(t, codec)
	codec = Segmented(
		Segmented(
			HybridSnappy(),
			8,
		),
		16,
	)
	testCodec(t, codec)
}

func TestHybridCompressed(t *testing.T) {
	codec := HybridCompressed(UseSnappy())
	testCodec(t, codec)
}

func TestSnappyStream(t *testing.T) {
	codec := HybridCompressed(UseSnappyStream())
	testCodec(t, codec)
}

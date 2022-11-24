// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

//go:build go1.16
// +build go1.16

package filebase

import (
	"bytes"
	"io"
	"math/rand"
	"testing"
	"testing/iotest"
	"time"

	"github.com/reusee/e5"
	"github.com/reusee/june/entity"
	"github.com/reusee/june/index"
	"github.com/reusee/pr2"
	"github.com/reusee/sb"
)

func TestContent(
	t *testing.T,
	scope Scope,
	wg *pr2.WaitGroup,
) {
	defer he(nil, e5.TestingFatal(t))

	chunkThreshold := ChunkThreshold(256)
	maxChunkSize := MaxChunkSize(1024)

	scope.Fork(&chunkThreshold, &maxChunkSize).Call(func(
		to ToContents,
		write WriteContents,
		newReader NewContentReader,
		selIndex index.SelectIndex,
		fetch entity.Fetch,
		del entity.Delete,
	) {

		testKeys := func(keys []Key, lengths []int64, bs []byte) {
			reader := newReader(keys, lengths)
			err := iotest.TestReader(reader, bs)
			ce(err)
			buf := new(bytes.Buffer)
			err = write(keys, buf)
			ce(err)
			if !bytes.Equal(bs, buf.Bytes()) {
				t.Fatal()
			}

			if reader.getLen() != int64(len(bs)) {
				t.Fatal()
			}

			// test Seek
			for _, i := range rand.Perm(len(bs)) {
				n, err := reader.Seek(int64(i), io.SeekStart)
				ce(err)
				if n != int64(i) {
					t.Fatal()
				}
				bs2, err := io.ReadAll(reader)
				ce(err)
				if !bytes.Equal(bs2, bs[i:]) {
					t.Fatal()
				}
			}

		}

		threshold := int64(chunkThreshold)
		max := int64(maxChunkSize)

		r := rand.New(rand.NewSource(time.Now().Unix()))

		for size := threshold; size <= max*2; size *= 2 {

			numBytes := 0

			n := 100
			for i := 0; i < n; i++ {
				bs := make([]byte, size)
				_, err := io.ReadFull(r, bs)
				ce(err)

				numBytes += len(bs)
				keys, lengths, err := to(wg, bytes.NewReader(bs), int64(len(bs)))
				ce(err)
				testKeys(keys, lengths, bs)
				var sum int64
				for _, l := range lengths {
					sum += l
				}
				if sum != int64(len(bs)) {
					t.Fatalf("expected %d, got %d", len(bs), sum)
				}

				bs = append(bs, []byte("foo")...)
				numBytes += len(bs)
				keys, lengths, err = to(wg, bytes.NewReader(bs), int64(len(bs)))
				ce(err)
				testKeys(keys, lengths, bs)

				bs = append([]byte("foo"), bs...)
				numBytes += len(bs)
				keys, lengths, err = to(wg, bytes.NewReader(bs), int64(len(bs)))
				ce(err)
				testKeys(keys, lengths, bs)

			}

			contentBytes := 0

			var contentKeys []Key
			selIndex(
				entity.MatchType(Content{}),
				index.TapKey(func(key Key) {
					var content Content
					ce(fetch(key, &content))
					contentBytes += len(content)
					contentKeys = append(contentKeys, key)
				}),
			)
			for _, key := range contentKeys {
				ce(del(key))
			}

			compactRatio := float64(contentBytes) / float64(numBytes)
			if compactRatio > 0.85 {
				t.Fatalf(
					"size %d, ratio %.3f",
					size,
					compactRatio,
				)
			}

		}
	})

}

func TestContentSB(t *testing.T) {
	src := rand.New(rand.NewSource(time.Now().Unix()))
	for i := 0; i < 8*1024; i++ {
		bs := make([]byte, i)
		_, err := io.ReadFull(src, bs)
		ce(err)
		a := Content(bs)
		var b Content
		ce(sb.Copy(
			sb.Marshal(a),
			sb.Unmarshal(&b),
		))
		if !bytes.Equal(a, b) {
			t.Fatal()
		}
	}
}

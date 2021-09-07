// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package codec

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"

	"github.com/reusee/june/opts"
	"github.com/reusee/sb"
)

type aeadCodec struct {
	newAEAD func() (cipher.AEAD, error)
	id      string
}

var _ Codec = aeadCodec{}

func NewAEADCodec(
	id string,
	newAEAD func() (cipher.AEAD, error),
) aeadCodec {
	return aeadCodec{
		id:      id,
		newAEAD: newAEAD,
	}
}

func (a aeadCodec) ID() string {
	return a.id
}

func (a aeadCodec) Encode(sink sb.Sink, options ...Option) sb.Sink {
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
		func(_ *sb.Token) (_ sb.Sink, err error) {
			defer he(&err)
			plaintext := buf.Bytes()
			aead, err := a.newAEAD()
			ce(err)
			nonce := make([]byte, aead.NonceSize())
			_, err = io.ReadFull(rand.Reader, nonce)
			ce(err)
			ciphertext := aead.Seal(nil, nonce, plaintext, nil)
			if err := sb.Copy(
				sb.Marshal(func() ([]byte, []byte) {
					return nonce, ciphertext
				}),
				sink,
			); err != nil {
				return nil, err
			}
			return nil, nil
		},
	)
}

func (a aeadCodec) Decode(src sb.Proc, options ...Option) sb.Proc {
	proc := sb.Proc(func() (_ *sb.Token, _ sb.Proc, err error) {
		defer he(&err)
		var nonce []byte
		var ciphertext []byte
		if err := sb.Copy(
			src,
			sb.Unmarshal(func(n []byte, c []byte) {
				nonce = n
				ciphertext = c
			}),
		); err != nil {
			return nil, nil, err
		}
		aead, err := a.newAEAD()
		ce(err)
		plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
		ce(err)
		br := bytes.NewReader(plaintext)
		buf := make([]byte, 8)
		return nil, sb.DecodeBuffer(br, br, buf, nil), nil
	})
	return proc
}

func AESGCM(key []byte) aeadCodec {
	newAEAD := func() (_ cipher.AEAD, err error) {
		defer he(&err)
		block, err := aes.NewCipher(key)
		ce(err)
		aead, err := cipher.NewGCM(block)
		ce(err)
		return aead, nil
	}
	var h []byte
	if err := sb.Copy(
		sb.Marshal(key),
		sb.Hash(sha256.New, &h, nil),
	); err != nil {
		panic(err)
	}
	return NewAEADCodec(
		fmt.Sprintf("aead-%x-", h),
		newAEAD,
	)
}

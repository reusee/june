// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package key

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"

	"github.com/reusee/e4"
)

const HashSize = 32

type Hash [HashSize]byte

func (h Hash) String() string {
	return hex.EncodeToString(h[:])
}

var emptyHash = Hash{}

func (h Hash) Valid() bool {
	return h != emptyHash
}

func HashFromString(str string) (hash Hash, err error) {
	defer he(&err)
	var bs []byte
	bs, err = hex.DecodeString(str)
	ce(err)
	if len(bs) > len(hash) {
		err = we(ErrTooLong, e4.With(fmt.Errorf("string: %s", str)))
		return
	}
	copy(hash[:], bs)
	return
}

type NewHashState func() hash.Hash

func (_ Def) NewHashState() NewHashState {
	return func() hash.Hash {
		return sha256.New()
	}
}

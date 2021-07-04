// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package key

import (
	"crypto/sha256"
	"hash"
)

type NewHashState func() hash.Hash

func (_ Def) NewHashState() NewHashState {
	return func() hash.Hash {
		return sha256.New()
	}
}

// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package key

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"testing"

	"github.com/reusee/e4"
)

func TestKeyJSON(t *testing.T) {
	defer he(nil, e4.TestingFatal(t))
	j, err := json.Marshal(Key{
		Namespace: Namespace{'a', 'b', 'c'},
		Hash:      Hash{1, 2, 3},
	})
	ce(err)
	if !bytes.Equal(
		j,
		[]byte(`"abc:0102030000000000000000000000000000000000000000000000000000000000"`),
	) {
		t.Fatal()
	}
}

func TestNamespaceFromString(t *testing.T) {
	defer he(nil, e4.TestingFatal(t))
	for i := 0; i < 128; i++ {
		var ns Namespace
		rand.Read(ns[:rand.Intn(len(ns))+1])
		str := ns.String()
		ns2, err := NamespaceFromString(str)
		ce(err)
		if ns != ns2 {
			t.Fatal()
		}
	}
}

func TestHashFromString(t *testing.T) {
	defer he(nil, e4.TestingFatal(t))
	for i := 0; i < 128; i++ {
		var hash Hash
		rand.Read(hash[:rand.Intn(len(hash))+1])
		str := hash.String()
		hash2, err := HashFromString(str)
		ce(err)
		if hash2 != hash {
			t.Fatal()
		}
	}
}

func TestKeyFromString(t *testing.T) {
	defer he(nil, e4.TestingFatal(t))
	for i := 0; i < 128; i++ {
		var key Key
	gen:
		rand.Read(key.Namespace[:rand.Intn(len(key.Namespace))+1])
		if bytes.Contains(key.Namespace[:], []byte(":")) {
			goto gen
		}
		rand.Read(key.Hash[:rand.Intn(len(key.Hash))+1])
		str := key.String()
		key2, err := KeyFromString(str)
		ce(err)
		if key2 != key {
			t.Fatal()
		}
	}
}

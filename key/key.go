// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package key

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/reusee/e4"
)

// namespace

type Namespace [8]byte

var emptyNamespace = Namespace{}

func (n Namespace) Valid() bool {
	return n != emptyNamespace
}

func (n Namespace) String() string {
	return fmt.Sprintf("%s", bytes.TrimRight(n[:], "\000"))
}

func NamespaceFromString(s string) (ns Namespace, err error) {
	if len(s) > len(ns) {
		err = we(ErrTooLong, e4.With(fmt.Errorf("string: %s", s)))
		return
	}
	copy(ns[:], []byte(s))
	return
}

// hash

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

// key

type Key struct {
	Namespace Namespace
	Hash      Hash
}

func (k Key) String() string {
	return k.Namespace.String() + ":" + k.Hash.String()
}

func (k Key) Valid() bool {
	return k.Namespace.Valid() && k.Hash.Valid()
}

var _ json.Marshaler = Key{}

func (k Key) MarshalJSON() ([]byte, error) {
	buf := new(bytes.Buffer)
	buf.WriteString(`"`)
	buf.WriteString(k.String())
	buf.WriteString(`"`)
	return buf.Bytes(), nil
}

var _ error = Key{}

func (k Key) Error() string {
	return fmt.Sprintf("key: %s", k.String())
}

func (k Key) Compare(key2 Key) int {
	if res := bytes.Compare(
		k.Namespace[:],
		key2.Namespace[:],
	); res != 0 {
		return res
	}
	if res := bytes.Compare(
		k.Hash[:],
		key2.Hash[:],
	); res != 0 {
		return res
	}
	return 0
}

func KeyFromString(str string) (key Key, err error) {
	defer he(&err)
	parts := strings.Split(str, ":")
	if len(parts) != 2 {
		err = we(ErrBadKey, e4.With(fmt.Errorf("key: %s", str)))
		return
	}
	var ns Namespace
	ns, err = NamespaceFromString(parts[0])
	ce(err)
	key.Namespace = ns
	var hash Hash
	hash, err = HashFromString(parts[1])
	ce(err)
	key.Hash = hash
	return
}

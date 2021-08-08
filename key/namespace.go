// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package key

import (
	"bytes"
	"fmt"

	"github.com/reusee/e4"
)

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

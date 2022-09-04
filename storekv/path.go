// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storekv

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"

	"github.com/reusee/e5"
	"github.com/reusee/june/key"
)

const objPrefix = "obj/"

func (s *Store) objPrefix() string {
	buf := new(strings.Builder)
	buf.WriteString(s.codec.ID())
	buf.WriteString(s.prefix)
	buf.WriteString(objPrefix)
	return buf.String()
}

func (s *Store) nsPrefix(ns key.Namespace) string {
	buf := new(strings.Builder)
	buf.WriteString(s.objPrefix())
	i := bytes.IndexByte(ns[:], '0')
	if i == -1 {
		buf.WriteString(hex.EncodeToString(ns[:]))
	} else {
		buf.WriteString(hex.EncodeToString(ns[:i]))
	}
	buf.WriteString("/")
	return buf.String()
}

func (s *Store) keyToPath(key Key) string {
	buf := new(strings.Builder)
	buf.WriteString(s.nsPrefix(key.Namespace))
	buf.WriteString(key.Hash.String())
	return buf.String()
}

var pathPattern = regexp.MustCompile(`([^)]+)/([0-9a-f]+)$`)

func (s *Store) pathToKey(path string) (key Key, err error) {
	defer he(&err, e5.Info("path %s", path))
	path = strings.TrimPrefix(path, s.codec.ID())
	path = strings.TrimPrefix(path, s.prefix)
	path = strings.TrimPrefix(path, objPrefix)
	parts := pathPattern.FindStringSubmatch(path)
	if len(parts) != 3 {
		ce(fmt.Errorf("bad path"))
	}
	bs, err := hex.DecodeString(parts[1])
	ce(err, e5.Info("bad namespace"))
	copy(key.Namespace[:], bs)
	bs, err = hex.DecodeString(parts[2])
	ce(err, e5.Info("bad key"))
	copy(key.Hash[:], bs)
	return
}

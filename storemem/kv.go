// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storemem

import (
	"bytes"
	"io"
	"strings"

	"github.com/reusee/e5"
	"github.com/reusee/june/storekv"
)

var _ storekv.KV = new(Store)

func (s *Store) CostInfo() storekv.CostInfo {
	return storekv.CostInfo{}
}

func (s *Store) KeyExists(key string) (bool, error) {
	select {
	case <-s.wg.Done():
		return false, ErrClosed
	default:
	}
	_, ok := s.values.Load(key)
	return ok, nil
}

func (s *Store) KeyGet(key string, fn func(r io.Reader) error) (err error) {
	select {
	case <-s.wg.Done():
		return ErrClosed
	default:
	}
	defer he(&err,
		e5.With(storekv.StringKey(key)),
	)
	v, ok := s.values.Load(key)
	if !ok {
		return we.With(e5.With(storekv.StringKey(key)))(ErrKeyNotFound)
	}
	if fn != nil {
		err := fn(bytes.NewReader(v.([]byte)))
		ce(err)
	}
	return nil
}

func (s *Store) KeyIter(prefix string, fn func(key string) error) error {
	select {
	case <-s.wg.Done():
		return ErrClosed
	default:
	}
	var err error
	s.values.Range(func(k, _ any) bool {
		key := k.(string)
		if !strings.HasPrefix(key, prefix) {
			return true
		}
		if err = fn(key); is(err, Break) {
			err = nil
			return false
		} else if err != nil {
			return false
		}
		return true
	})
	return err
}

func (s *Store) KeyPut(key string, r io.Reader) error {
	select {
	case <-s.wg.Done():
		return ErrClosed
	default:
	}
	// do not retain reader bytes
	bs, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	s.values.Store(key, bs)
	return nil
}

func (s *Store) KeyDelete(keys ...string) error {
	select {
	case <-s.wg.Done():
		return ErrClosed
	default:
	}
	for _, key := range keys {
		s.values.Delete(key)
	}
	return nil
}

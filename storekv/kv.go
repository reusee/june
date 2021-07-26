// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storekv

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/reusee/june/key"
	"github.com/reusee/june/store"
	"github.com/reusee/june/sys"
	"github.com/reusee/pr"
	"github.com/reusee/sb"
)

type StringKey string

func (s StringKey) Error() string {
	return "key: " + string(s)
}

type KV interface {
	StoreID() string                      // IndexID is only used in initialization, should be stable and unique
	Name() string                         // Name is for human readable
	KeyPut(key string, r io.Reader) error // implementation should not retain reader underlying bytes after return
	KeyGet(key string, fn func(io.Reader) error) error
	KeyExists(key string) (bool, error)
	KeyIter(prefix string, fn func(key string) error) error
	KeyDelete(key ...string) error
}

type Store struct {
	*pr.WaitTree
	name          string
	kv            KV
	codec         Codec
	cache         store.Cache
	newHashState  key.NewHashState
	_id           store.ID
	prefix        string
	setupIDOnce   sync.Once
	readDisabled  bool
	writeDisabled bool
	parallel      int
	offloads      []WithOffload
}

var _ store.Store = new(Store)

type New func(
	kv KV,
	prefix string,
	options ...NewOption,
) (*Store, error)

type NewOption interface {
	IsNewOption()
}

var serial int64

func (_ Def) New(
	newHashState key.NewHashState,
	parallel sys.Parallel,
	wt *pr.WaitTree,
) (
	newStore New,
) {

	newStore = func(
		kv KV,
		prefix string,
		options ...NewOption,
	) (*Store, error) {

		codec := Codec(DefaultCodec)
		var cache Cache
		writeDisabled := false
		readDisabled := false
		var offloads []WithOffload

		for _, option := range options {
			switch option := option.(type) {
			case withCodec:
				codec = option.Codec
			case withCache:
				cache = option.Cache
			case WithoutWrite:
				writeDisabled = true
			case WithoutRead:
				readDisabled = true
			case WithOffload:
				offloads = append(offloads, option)
			default:
				panic(fmt.Errorf("not handled option: %T", option))
			}
		}

		store := &Store{
			WaitTree: wt,
			name: fmt.Sprintf("kv%d(%s, %s)",
				atomic.AddInt64(&serial, 1),
				kv.Name(),
				prefix,
			),
			kv:            kv,
			newHashState:  newHashState,
			prefix:        prefix,
			codec:         codec,
			cache:         cache,
			writeDisabled: writeDisabled,
			readDisabled:  readDisabled,
			parallel:      int(parallel),
			offloads:      offloads,
		}

		return store, nil
	}

	return
}

func (s *Store) Name() string {
	return s.name
}

func (s *Store) getID() (id store.ID, err error) {
	defer he(&err)
	s.setupIDOnce.Do(func() {
		idPath := strings.Join([]string{
			s.codec.ID(),
			s.prefix,
			"__id__",
		}, "")
		setupID := func() (err error) {
			defer he(&err)
			id := StoreID(fmt.Sprintf("kv(%s)", s.kv.StoreID()))
			buf := new(bytes.Buffer)
			if err := sb.Copy(
				sb.Marshal(id),
				sb.Encode(buf),
			); err != nil {
				return err
			}
			err = s.kv.KeyPut(idPath, buf)
			ce(err)
			s._id = id
			return nil
		}

		if err := s.kv.KeyGet(idPath, func(r io.Reader) (err error) {
			defer he(&err)
			var id StoreID
			if err := sb.Copy(
				sb.Decode(r),
				sb.Unmarshal(&id),
			); err != nil {
				return err
			}
			s._id = id
			return nil
		}); is(err, ErrKeyNotFound) {
			err := setupID()
			ce(err)
		} else {
			ce(err)
		}
	})

	id = s._id
	return
}

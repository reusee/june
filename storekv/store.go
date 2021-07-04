// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storekv

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"runtime"

	"github.com/reusee/e4"
	"github.com/reusee/june/key"
	"github.com/reusee/june/opts"
	"github.com/reusee/june/store"
	"github.com/reusee/pr"
	"github.com/reusee/sb"
)

func (s *Store) ID() (StoreID, error) {
	return s.getID()
}

func (s *Store) Close() error {
	return s.kv.Close()
}

func (s *Store) Exists(key Key) (bool, error) {
	path := s.keyToPath(key)
	return s.kv.KeyExists(path)
}

func (s *Store) IterAllKeys(fn func(Key) error) error {
	return s.kv.KeyIter(s.objPrefix(), func(k string) error {
		key, err := s.pathToKey(k)
		if err != nil {
			// ignore
			return nil
		}
		return fn(key)
	})
}

var shards = func() []string {
	var ret []string
	for i := 0; i < 256; i++ {
		ret = append(ret, fmt.Sprintf("%02x", i))
	}
	return ret
}()

func (s *Store) IterKeys(ns key.Namespace, fn func(Key) error) error {
	nsPrefix := s.nsPrefix(ns)

	ctx, cancel := context.WithCancel(s.ctx)
	defer cancel()

	keys := make(chan *Key)
	errCh := make(chan error, 1)
	setErr := func(err error) {
		select {
		case errCh <- err:
		default:
		}
	}

	go func() {
		sem := make(chan struct{}, s.parallel)
	loop:
		for _, shard := range shards {
			shard := shard

			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				break loop
			}
			go func() {
				defer func() {
					<-sem
				}()

				prefix := nsPrefix + shard
				if err := s.kv.KeyIter(prefix, func(k string) error {
					key, err := s.pathToKey(k)
					if err != nil {
						// ignore
						return nil
					}
					if key.Namespace != ns {
						return nil
					}
					select {
					case keys <- &key:
					case <-ctx.Done():
						return Break
					}
					return nil
				}); err != nil {
					setErr(err)
					return
				}

			}()

		}

		for i := 0; i < cap(sem); i++ {
			sem <- struct{}{}
		}
		close(keys)

	}()

	defer func() {
		// wait close
		<-keys
	}()

loop:
	for {
		select {

		case key := <-keys:
			if key == nil {
				break loop
			}
			if err := fn(*key); is(err, store.Break) {
				break loop
			} else if err != nil {
				return err
			}

		case err := <-errCh:
			return err

		}
	}

	select {
	case err := <-errCh:
		return err
	default:
	}

	return nil
}

func (s *Store) Read(key Key, fn func(sb.Stream) error) error {

	// cache
	if s.cache != nil {
		if err := s.cache.CacheGet(key, func(stream sb.Stream) (err error) {
			defer he(&err)
			// cache hit, check exists
			exists, err := s.Exists(key)
			ce(err)
			if !exists {
				return we(ErrKeyNotFound, e4.With(key))
			}
			err = fn(stream)
			ce(err, e4.With(ErrRead), e4.With(key))
			return nil
		}); is(err, ErrKeyNotFound) {
			// skip
		} else if err != nil {
			return err
		} else {
			return nil
		}
	}

	path := s.keyToPath(key)
	return s.kv.KeyGet(path, func(r io.Reader) (err error) {
		defer he(&err)
		var tokens sb.Tokens
		var sum []byte
		if err := sb.Copy(
			s.codec.Decode(sb.Decode(r)),
			sb.CollectTokens(&tokens),
			sb.Hash(s.newHashState, &sum, nil),
		); err != nil {
			return err
		}
		if !bytes.Equal(key.Hash[:], sum) {
			return we(ErrKeyNotMatch, e4.With(key))
		}
		err = fn(tokens.Iter())
		ce(err, e4.With(ErrRead), e4.With(key))
		return nil
	})

}

var bytesBufferPool = pr.NewPool(
	int32(runtime.NumCPU()),
	func() any {
		return new(bytes.Buffer)
	},
)

func (s *Store) Write(
	ns key.Namespace,
	stream sb.Stream,
	options ...WriteOption,
) (res store.WriteResult, err error) {
	defer he(&err)

	var tapKey TapKey
	var tapResult TapWriteResult
	var newBytesBuffer opts.NewBytesBuffer
	for _, option := range options {
		switch option := option.(type) {
		case TapKey:
			tapKey = option
		case TapWriteResult:
			tapResult = option
		case opts.NewBytesBuffer:
			newBytesBuffer = option
		}
	}

	defer func() {
		if err == nil && tapResult != nil {
			tapResult(res)
		}
	}()

	var hash []byte
	var tokens sb.Tokens
	if err = sb.Copy(
		stream,
		sb.Hash(s.newHashState, &hash, nil),
		sb.CollectTokens(&tokens),
	); err != nil {
		return
	}
	copy(res.Key.Hash[:], hash)
	res.Key.Namespace = ns

	if tapKey != nil {
		tapKey(res.Key)
	}

	// put cache
	if s.cache != nil && len(tokens) > 0 {
		err := s.cache.CachePut(res.Key, tokens)
		ce(err)
	}

	path := s.keyToPath(res.Key)
	var ok bool
	ok, err = s.kv.KeyExists(path)
	ce(err)
	if ok {
		return
	}

	v, put := bytesBufferPool.Get()
	buf := v.(*bytes.Buffer)
	defer func() {
		buf.Reset()
		put()
	}()
	var sink sb.Sink
	if newBytesBuffer != nil {
		sink = s.codec.Encode(
			sb.Encode(buf),
			newBytesBuffer,
		)
	} else {
		sink = s.codec.Encode(
			sb.Encode(buf),
		)
	}
	if err = sb.Copy(
		tokens.Iter(),
		sink,
	); err != nil {
		return
	}
	res.BytesWritten += int64(buf.Len())
	err = s.kv.KeyPut(path, buf)
	ce(err)

	res.Written = true

	return
}

func (s *Store) Delete(keys []Key) error {
	var paths []string
	for _, key := range keys {
		paths = append(paths, s.keyToPath(key))
	}
	return s.kv.KeyDelete(paths...)
}

func (s *Store) Sync() error {
	return s.kv.Sync()
}

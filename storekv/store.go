// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storekv

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/reusee/e5"
	"github.com/reusee/june/key"
	"github.com/reusee/june/opts"
	"github.com/reusee/june/store"
	"github.com/reusee/sb"
)

func (s *Store) ID() (StoreID, error) {
	return s.getID()
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

	ctx, cancel := context.WithCancel(s.wg)
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
				return we.With(e5.With(key))(ErrKeyNotFound)
			}
			err = fn(stream)
			ce(err, e5.With(ErrRead), e5.With(key))
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

		if len(s.offloads) > 0 {
			// offload
			if err := sb.Copy(
				sb.Deref(
					s.codec.Decode(sb.Decode(r)),
					func(value []byte) (sb.Stream, error) {
						if !bytes.Equal(value, []byte("offloaded")) {
							return nil, nil
						}
						for _, offload := range s.offloads {
							offloadStore := offload(key, -1)
							if offloadStore == nil {
								continue
							}
							var offloadTokens sb.Tokens
							err := offloadStore.Read(key, func(s sb.Stream) error {
								return sb.Copy(
									s,
									sb.CollectTokens(&offloadTokens),
								)
							})
							if is(err, ErrKeyNotFound) {
								continue
							}
							ce(err)
							return offloadTokens.Iter(), nil
						}
						return nil, nil
					},
				),
				sb.CollectTokens(&tokens),
				sb.Hash(s.newHashState, &sum, nil),
			); err != nil {
				return err
			}

		} else {
			if err := sb.Copy(
				s.codec.Decode(sb.Decode(r)),
				sb.CollectTokens(&tokens),
				sb.Hash(s.newHashState, &sum, nil),
			); err != nil {
				return err
			}
		}

		if !bytes.Equal(key.Hash[:], sum) {
			return we.With(e5.With(key))(ErrKeyNotMatch)
		}
		err = fn(tokens.Iter())
		ce(err, e5.With(ErrRead), e5.With(key))
		return nil
	})

}

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
	var encodedLen int
	if err = sb.Copy(
		stream,
		sb.Hash(s.newHashState, &hash, nil),
		sb.CollectTokens(&tokens),
		sb.EncodedLen(&encodedLen, nil),
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
		err := s.cache.CachePut(res.Key, tokens, store.EncodedLen(encodedLen))
		ce(err)
	}

	path := s.keyToPath(res.Key)

	if s.costInfo.Exists <= s.costInfo.Put {
		var ok bool
		ok, err = s.kv.KeyExists(path)
		ce(err)
		if ok {
			return
		}
	}

	// offload
	offloaded := false
	if len(s.offloads) > 0 {
		for _, offload := range s.offloads {
			offloadStore := offload(res.Key, encodedLen)
			if offloadStore == nil {
				continue
			}
			offloaded = true
			result, e := offloadStore.Write(ns, tokens.Iter(), options...)
			ce(e)
			if result.Key != res.Key {
				err = we(ErrKeyNotMatch)
				return
			}
		}
	}

	buf := new(bytes.Buffer)
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

	var src sb.Stream
	if offloaded {
		src = sb.Tokens{
			{
				Kind:  sb.KindRef,
				Value: []byte("offloaded"),
			},
		}.Iter()
	} else {
		src = tokens.Iter()
	}

	if err = sb.Copy(
		src,
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

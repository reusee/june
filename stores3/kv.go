// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package stores3

import (
	"bytes"
	"context"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/reusee/e5"
	"github.com/reusee/june/storekv"
)

func (k *KV) KeyExists(key string) (_ bool, err error) {
	defer k.wg.Add()()
	defer he(&err)
	ctx, cancel := context.WithTimeout(k.wg, k.timeout)
	defer cancel()
	_, err = k.client.StatObject(ctx, k.bucket, key, minio.StatObjectOptions{})
	if err != nil {
		var resp minio.ErrorResponse
		if as(err, &resp) && resp.Code == "NoSuchKey" {
			return false, nil
		}
		ce(err)
	}
	return true, nil
}

func (k *KV) KeyGet(key string, fn func(io.Reader) error) (err error) {
	defer k.wg.Add()()
	defer he(&err,
		e5.With(storekv.StringKey(key)),
	)
	ctx, cancel := context.WithTimeout(k.wg, k.timeout)
	defer cancel()
	obj, err := k.client.GetObject(ctx, k.bucket, key, minio.GetObjectOptions{})
	var resp minio.ErrorResponse
	if as(err, &resp) && resp.Code == "NoSuchKey" {
		return we.With(e5.With(storekv.StringKey(key)))(ErrKeyNotFound)
	}
	ce(err)
	defer obj.Close()
	if fn != nil {
		if err := fn(obj); as(err, &resp) && resp.Code == "NoSuchKey" {
			return we.With(e5.With(storekv.StringKey(key)))(ErrKeyNotFound)
		} else {
			ce(err)
		}
	}
	return nil
}

func (k *KV) KeyPut(key string, r io.Reader) (err error) {
	defer k.wg.Add()()
	defer he(&err,
		e5.With(storekv.StringKey(key)),
	)
	ctx, cancel := context.WithTimeout(k.wg, k.timeout)
	defer cancel()
	var content []byte
	if b, ok := r.(interface {
		Bytes() []byte
	}); ok {
		content = b.Bytes()
	} else {
		var err error
		content, err = io.ReadAll(r)
		ce(err)
	}
	if _, err := k.client.PutObject(
		ctx, k.bucket, key,
		bytes.NewReader(content), int64(len(content)),
		minio.PutObjectOptions{},
	); err != nil {
		return err
	}
	return nil
}

func (k *KV) KeyIter(prefix string, fn func(string) error) (err error) {
	defer k.wg.Add()()
	defer he(&err, e5.Info("prefix %s", prefix))

	marker := ""
loop:
	for {
		select {
		case <-k.wg.Done():
			break loop
		default:
		}
		res, err := k.core.ListObjects(
			k.bucket, prefix, marker, "", -1,
		)
		ce(err)
		if len(res.Contents) == 0 {
			break
		}
		for _, info := range res.Contents {
			if err := fn(info.Key); is(err, Break) {
				return nil
			} else {
				ce(err)
			}
		}
		if !res.IsTruncated {
			break
		}
		marker = res.Contents[len(res.Contents)-1].Key
	}

	return nil
}

func (k *KV) KeyDelete(keys ...string) (err error) {
	defer k.wg.Add()()
	defer he(&err)
	for len(keys) > 0 {
		i := 0

		ctx, cancel := context.WithTimeout(k.wg, k.timeout)
		defer cancel()
		ch := make(chan minio.ObjectInfo)
		errChan := k.client.RemoveObjects(
			ctx,
			k.bucket,
			ch,
			minio.RemoveObjectsOptions{},
		)
		for ; i < 1000 && i < len(keys); i++ {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case err := <-errChan:
				return err.Err
			case ch <- minio.ObjectInfo{
				Key: keys[i],
			}:
			}
		}
		close(ch)
		err := <-errChan
		ce(err.Err)

		keys = keys[i:]
	}
	return nil
}

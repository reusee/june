// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package stores3

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/reusee/june/storekv"
	"github.com/reusee/pr"
)

type KV struct {
	*pr.WaitTree
	name    string
	storeID string

	client  *minio.Client
	core    *minio.Core
	timeout time.Duration

	endpoint string
	key      string
	secret   string
	bucket   string

	closeOnce sync.Once
}

var _ storekv.KV = new(KV)

type New func(
	endpoint string,
	key string,
	secret string,
	useSSL bool,
	bucket string,
	options ...NewOption,
) (*KV, error)

type NewOption interface {
	IsNewOption()
}

func (_ Def) New(
	timeout Timeout,
	parentWt *pr.WaitTree,
) New {
	return func(
		endpoint string,
		key string,
		secret string,
		useSSL bool,
		bucket string,
		options ...NewOption,
	) (_ *KV, err error) {
		defer he(&err)

		client, err := minio.New(endpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(key, secret, ""),
			Secure: true,
		})
		ce(err)

		core, err := minio.NewCore(endpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(key, secret, ""),
			Secure: true,
		})
		ce(err)

		kv := &KV{
			WaitTree: parentWt,
			name: fmt.Sprintf("s3%d(%s)",
				atomic.AddInt64(&serial, 1),
				bucket,
			),
			storeID: fmt.Sprintf("s3(%s, %s)",
				endpoint,
				bucket,
			),
			client:   client,
			core:     core,
			bucket:   bucket,
			timeout:  time.Duration(timeout),
			endpoint: endpoint,
			key:      key,
			secret:   secret,
		}

		return kv, nil
	}
}

var serial int64

func (k *KV) Name() string {
	return k.name
}

func (k *KV) StoreID() string {
	return k.storeID
}

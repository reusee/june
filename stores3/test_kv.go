// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package stores3

import (
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/pelletier/go-toml"
	"github.com/reusee/e4"
	"github.com/reusee/june/storekv"
)

func TestKV(
	t *testing.T,
	testKV storekv.TestKV,
	newKV New,
) {
	defer he(nil, e4.TestingFatal(t))

	configDir, err := os.UserConfigDir()
	ce(err)
	content, err := os.ReadFile(filepath.Join(configDir, "qcloud-cos-key.toml"))
	ce(err)
	var config struct {
		Endpoint   string
		TestBucket string
		Key        string
		Secret     string
	}
	err = toml.Unmarshal(content, &config)
	ce(err)

	with := func(fn func(storekv.KV, string)) {
		kv, err := newKV(
			config.Endpoint,
			config.Key,
			config.Secret,
			true,
			config.TestBucket,
		)
		ce(err)
		defer kv.Close()
		fn(kv, strconv.FormatInt(rand.Int63(), 10))
	}
	testKV(t, with)
}

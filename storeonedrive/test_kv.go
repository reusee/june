// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storeonedrive

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/reusee/e4"
	"github.com/reusee/june/storekv"
	"github.com/reusee/pr"
	"github.com/reusee/sb"
	"golang.org/x/oauth2"
)

func TestKV(
	t *testing.T,
	wt *pr.WaitTree,
	testKV storekv.TestKV,
	newStore New,
	ctx context.Context,
) {

	if os.Getenv("test_onedrive") == "" {
		t.Skip()
	}

	defer he(nil, e4.TestingFatal(t))

	with := func(fn func(storekv.KV, string)) {
		config := oauth2.Config{
			ClientID: "c6937f2a-2038-46b5-85ab-ccc9f1d60eef",
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
				TokenURL: "https://login.microsoftonline.com/common/oauth2/v2.0/token",
			},
			Scopes: []string{
				"Files.ReadWrite.AppFolder",
				"offline_access",
			},
		}

		var token *oauth2.Token
		configDir, err := os.UserConfigDir()
		ce(err)
		content, err := os.ReadFile(
			filepath.Join(configDir, "test-onedrive.token"),
		)
		ce(err)
		ce(sb.Copy(
			sb.Decode(bytes.NewReader(content)),
			sb.Unmarshal(&token),
		))

		client := config.Client(ctx, token)
		dir := fmt.Sprintf("%d", rand.Int63())
		kv, err := newStore(
			wt,
			client,
			"/me/drive/special/AppRoot/",
			dir,
		)
		ce(err)
		fn(kv, "foo")

	}

	testKV(t, with)

}

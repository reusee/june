// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

//go:build ignore
// +build ignore

package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/reusee/e4"
	"github.com/reusee/ling/v2/oauth"
	"github.com/reusee/sb"
	"golang.org/x/oauth2"
	"os"
	"path/filepath"
	"time"
)

func main() {

	config := &oauth2.Config{
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

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*16)
	defer cancel()
	token, err := oauth.GetToken(
		ctx,
		config,
		func(url string) {
			pt("%s\n", url)
		},
	)
	ce(err)

	buf := new(bytes.Buffer)
	ce(sb.Copy(
		sb.Marshal(token),
		sb.Encode(buf),
	))
	configDir, err := os.UserConfigDir()
	ce(err)
	ce(os.WriteFile(
		filepath.Join(configDir, "ling-test-onedrive.token"),
		buf.Bytes(),
		0644,
	))

}

var (
	ce = e4.Check
	pt = fmt.Printf
)

// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package oauth

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"time"

	"golang.org/x/oauth2"
)

func GetToken(
	ctx context.Context,
	config *oauth2.Config,
	showURL func(string),
) (
	token *oauth2.Token,
	err error,
) {
	defer he(&err)

	state := fmt.Sprintf("%d-%d", time.Now().UnixNano(), rand.Intn(1<<31))
	codeCh := make(chan string)

	originRedirectURL := config.RedirectURL
	defer func() {
		config.RedirectURL = originRedirectURL
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		query := req.URL.Query()
		reqState := query.Get("state")
		if reqState != state {
			return
		}
		code := query.Get("code")
		if code == "" {
			return
		}
		select {
		case codeCh <- code:
		default:
		}
	})
	port := rand.Intn(30000) + 10000
	for {
		addr := fmt.Sprintf("127.0.0.1:%d", port)
		ln, err := net.Listen("tcp", addr)
		if err != nil {
			port = rand.Intn(30000) + 10000
			continue
		}
		config.RedirectURL = fmt.Sprintf("http://localhost:%d", port)
		server := &http.Server{
			Handler: mux,
		}
		defer server.Close()
		go server.Serve(ln)
		break
	}

	authURL := config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	showURL(authURL)

	var code string
	select {
	case <-ctx.Done():
		err = ctx.Err()
		return
	case code = <-codeCh:
	}
	ctx, cancel := context.WithTimeout(ctx, time.Second*32)
	defer cancel()
	token, err = config.Exchange(ctx, code, oauth2.AccessTypeOffline)
	ce(err)

	return
}

// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storeonedrive

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/reusee/e5"
	"github.com/reusee/pr"
)

type New func(
	parentWt *pr.WaitTree,
	client *http.Client,
	drivePath string,
	dir string,
) (*Store, error)

type Store struct {
	*pr.WaitTree
	name      string
	storeID   string
	dirOK     sync.Map
	client    *http.Client
	drivePath string
	dir       string
	idByPath  sync.Map
}

func (_ Def) New() New {
	return func(
		parentWt *pr.WaitTree,
		client *http.Client,
		drivePath string,
		dir string,
	) (
		drive *Store,
		err error,
	) {

		return &Store{
			WaitTree: parentWt,
			name: fmt.Sprintf("onedrive%d(%s)",
				atomic.AddInt64(&serial, 1),
				dir,
			),
			storeID: fmt.Sprintf("onedrive(%s, %s)",
				drivePath,
				dir,
			),
			client:    client,
			drivePath: path.Clean(drivePath),
			dir:       dir,
		}, nil
	}
}

var serial int64

func (s *Store) Name() string {
	return s.name
}

func (s *Store) StoreID() string {
	return s.storeID
}

var ErrNotFound = errors.New("not found")

var ErrExisted = errors.New("existed")

func (s *Store) req(
	ctx context.Context,
	method string,
	path string,
	body io.Reader,
	contentType string,
	target any,
) (err error) {
	select {
	case <-s.Ctx.Done():
		return ErrClosed
	default:
	}
	defer he(&err, e5.Info("request %s %s", method, path))

	resp, err := s.request(ctx, method, path, body, contentType)
	ce(err)
	defer resp.Body.Close()
	content, err := io.ReadAll(resp.Body)
	ce(err)
	var data struct {
		Error *struct {
			Message string
		}
	}
	if len(content) > 0 {
		err := json.Unmarshal(content, &data)
		ce(err, e5.Info("json %s", content))
	}
	if data.Error != nil {
		switch resp.StatusCode {
		case 404:
			return fmt.Errorf("%s: %w", content, ErrNotFound)
		case 409:
			return fmt.Errorf("%s: %w", content, ErrExisted)
		case 429:
			// request throttle
			pt("%s\n", content)
			time.Sleep(time.Minute) //TODO use duration from response
		}
		ce(fmt.Errorf("%s", content))
	}
	if target != nil {
		err := json.Unmarshal(content, target)
		ce(err, e5.Info("json %s", content))
	}
	return nil
}

func (s *Store) request(
	ctx context.Context,
	method string,
	path string,
	body io.Reader,
	contentType string,
) (_ *http.Response, err error) {
	select {
	case <-s.Ctx.Done():
		return nil, ErrClosed
	default:
	}
	defer he(&err)

	if !strings.HasPrefix(path, "https") {
		path = "https://graph.microsoft.com/v1.0" + path
	}
	req, err := http.NewRequestWithContext(
		ctx,
		method,
		path,
		body,
	)
	ce(err)
	if contentType != "" {
		req.Header.Add("Content-Type", contentType)
	}
	resp, err := s.client.Do(req)
	ce(err)
	return resp, nil
}

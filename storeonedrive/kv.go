// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package storeonedrive

import (
	"context"
	"io"
	"path"
	"strings"
	"time"

	"github.com/reusee/e5"
	"github.com/reusee/june/storekv"
)

const (
	defaultTimeout = time.Hour * 32
)

var _ storekv.KV = new(Store)

func (s *Store) CostInfo() storekv.CostInfo {
	return storekv.CostInfo{}
}

func (s *Store) keyToShardedRelPath(key string) string {
	parts := strings.Split(key, "/")
	hex := parts[len(parts)-1]
	parts = append(
		parts[:len(parts)-1],
		hex[:2],
		hex,
	)
	return path.Clean(path.Join(parts...))
}

func (s *Store) keyToDrivePath(key string, sub string) string {
	rel := s.keyToShardedRelPath(key)
	return s.relToDrivePath(rel, sub)
}

func (s *Store) relToDrivePath(rel string, sub string) string {
	rel = strings.Map(func(r rune) rune {
		switch r {
		case '"', '*', ':', '<', '>', '?', '\\', '|', '~', '#', '%', '&', '{', '}':
			return '='
		}
		return r
	}, rel)
	return path.Clean(path.Join(
		s.drivePath+":",
		path.Join(s.dir, rel)+":",
		sub,
	))
}

func (s *Store) shardedRelPathToKey(p string) string {
	parts := strings.Split(p, "/")
	parts = append(
		parts[:len(parts)-2],
		parts[len(parts)-1],
	)
	return path.Join(parts...)
}

func (s *Store) KeyDelete(ctx context.Context, keys ...string) (err error) {
	select {
	case <-ctx.Done():
		return ErrClosed
	default:
	}
	defer he(&err)
	for _, key := range keys {
		path := s.keyToDrivePath(key, "")
		ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
		err := s.req(ctx, "DELETE", path, nil, "", nil)
		cancel()
		ce(err, e5.Info("delete %s %s", key, path))
	}
	return nil
}

func (s *Store) KeyExists(ctx context.Context, key string) (ok bool, err error) {
	select {
	case <-ctx.Done():
		return false, ErrClosed
	default:
	}
	defer he(&err,
		e5.With(storekv.StringKey(key)),
	)
	path := s.keyToDrivePath(key, "")
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()
	err = s.req(ctx, "GET", path, nil, "", nil)
	if is(err, ErrNotFound) {
		return false, nil
	}
	ce(err, e5.Info("path %s", path))
	return true, nil
}

func (s *Store) KeyGet(ctx context.Context, key string, fn func(io.Reader) error) (err error) {
	select {
	case <-ctx.Done():
		return ErrClosed
	default:
	}
	defer he(&err,
		e5.With(storekv.StringKey(key)),
	)
	path := s.keyToDrivePath(key, "content")
	resp, err := s.request(ctx, "GET", path, nil, "")
	ce(err, e5.Info("path %s", path))
	defer resp.Body.Close()
	if resp.StatusCode == 404 {
		return we.With(e5.With(storekv.StringKey(key)))(storekv.ErrKeyNotFound)
	}
	return fn(resp.Body)
}

func (s *Store) iterFiles(
	ctx context.Context,
	dir string,
	fn func(path string, isDir bool) error,
) (err error) {
	select {
	case <-ctx.Done():
		return ErrClosed
	default:
	}
	defer he(&err, e5.Info("dir %s", dir))

	p := s.relToDrivePath(dir, "children")

do:
	var data struct {
		Next  string `json:"@odata.nextLink"`
		Value []struct {
			Folder *struct{}
			Name   string
		}
	}
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	err = s.req(ctx, "GET", p, nil, "", &data)
	cancel()
	if is(err, ErrNotFound) {
		return nil
	}
	ce(err, e5.Info("path %s", p))

	for _, row := range data.Value {
		isDir := row.Folder != nil
		filePath := path.Join(dir, row.Name)
		if err := fn(
			filePath,
			isDir,
		); err != nil {
			return err
		}
		if isDir {
			err := s.iterFiles(ctx, filePath, fn)
			ce(err)
		}
	}

	if data.Next != "" {
		p = data.Next
		goto do
	}

	return nil
}

func (s *Store) KeyIter(ctx context.Context, prefix string, fn func(string) error) (err error) {
	select {
	case <-ctx.Done():
		return ErrClosed
	default:
	}
	return s.iterFiles(ctx, prefix, func(filePath string, isDir bool) (err error) {
		defer he(&err)
		if isDir {
			return nil
		}
		key := s.shardedRelPathToKey(filePath)
		err = fn(key)
		ce(err, e5.Info("key %s", key))
		return nil
	})
}

func (s *Store) ensureDir(ctx context.Context, dir string) (err error) {
	defer he(&err)

	if dir == "/" || dir == "." {
		return nil
	}
	if _, ok := s.dirOK.Load(dir); ok {
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	err = s.req(ctx, "GET", s.relToDrivePath(dir, ""), nil, "", nil)
	cancel()
	if err == nil {
		s.dirOK.Store(dir, struct{}{})
		return nil
	}
	if !is(err, ErrNotFound) {
		return err
	}

	// create
	parent := path.Dir(dir)
	err = s.ensureDir(ctx, parent)
	ce(err)
	var addr string
	v, ok := s.idByPath.Load(parent)
	if ok {
		addr = "/me/drive/items/" + v.(string) + "/children"
	} else {
		addr = s.relToDrivePath(parent, "children")
	}
	var data struct {
		ID string
	}
	ctx, cancel = context.WithTimeout(ctx, defaultTimeout)
	err = s.req(
		ctx,
		"POST", addr,
		strings.NewReader(`{
        "name": "`+path.Base(dir)+`",
        "folder": {},
        "@microsoft.graph.conflictBehavior": "fail"
      }`),
		"application/json",
		&data,
	)
	cancel()
	if is(err, ErrExisted) {
		return nil
	}
	ce(err, e5.Info("create", parent))
	s.idByPath.Store(dir, data.ID)
	s.dirOK.Store(dir, struct{}{})

	return nil
}

func (s *Store) KeyPut(ctx context.Context, key string, r io.Reader) (err error) {
	select {
	case <-ctx.Done():
		return ErrClosed
	default:
	}
	defer he(&err,
		e5.With(storekv.StringKey(key)),
	)
	if err := s.ensureDir(
		ctx,
		path.Dir(s.keyToShardedRelPath(key)),
	); err != nil {
		return err
	}
	if err := s.req(
		ctx,
		"PUT", s.keyToDrivePath(key, "content"),
		r, "application/octet-stream",
		nil,
	); err != nil {
		return err
	}
	return nil
}

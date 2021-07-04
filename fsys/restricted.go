// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package fsys

import (
	"path/filepath"
	"sync"
)

type SetRestrictedPath func(string) error

type IsRestrictedPath func(string) (bool, error)

func (_ Def) SetRestrictedPath() (
	set SetRestrictedPath,
	is IsRestrictedPath,
) {

	type Node struct {
		Subs       map[string]*Node
		Restricted bool
	}
	newNode := func() *Node {
		return &Node{
			Subs: make(map[string]*Node),
		}
	}
	var l sync.RWMutex
	roots := make(map[string]*Node)

	var findNode func(string) (*Node, bool)
	findNode = func(path string) (*Node, bool) {
		dir, name := filepath.Split(path)
		if dir == path {
			// is root path
			root, ok := roots[dir]
			if !ok {
				root = newNode()
				roots[dir] = root
			}
			return root, false
		}
		dir = filepath.Clean(dir)
		parent, restricted := findNode(dir)
		if restricted {
			return nil, true
		}
		if parent == nil {
			return nil, false
		}
		node, ok := parent.Subs[name]
		if !ok {
			return nil, false
		}
		if node == nil {
			return nil, false
		}
		return node, node.Restricted
	}

	var setNode func(string) *Node
	setNode = func(path string) *Node {
		dir, name := filepath.Split(path)
		if dir == path {
			// is root path
			root, ok := roots[dir]
			if !ok {
				root = newNode()
				roots[dir] = root
			}
			return root
		}
		dir = filepath.Clean(dir)
		parent := setNode(dir)
		if node, ok := parent.Subs[name]; ok {
			return node
		}
		node := newNode()
		parent.Subs[name] = node
		return node
	}

	set = func(path string) (err error) {
		defer he(&err)
		l.Lock()
		defer l.Unlock()
		path, err = RealPath(path)
		ce(err)
		node := setNode(path)
		node.Restricted = true
		return nil
	}

	is = func(path string) (restricted bool, err error) {
		defer he(&err)
		l.RLock()
		defer l.RUnlock()
		path, err = RealPath(path)
		ce(err)
		_, restricted = findNode(path)
		return
	}

	return
}

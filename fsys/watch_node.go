// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package fsys

import (
	"strings"
	"time"
)

type Node struct {
	ModTime time.Time
	Subs    map[string]*Node
}

func NewTree(subPath []string, t time.Time) *Node {
	node := &Node{
		ModTime: t,
	}
	if len(subPath) > 0 {
		subNode := NewTree(subPath[1:], t)
		subName := subPath[0]
		node.Subs = map[string]*Node{
			subName: subNode,
		}
	}
	return node
}

func (n *Node) Update(subPath []string, t time.Time) {
	if t.After(n.ModTime) {
		n.ModTime = t
	}
	if len(subPath) == 0 {
		return
	}
	if n.Subs == nil {
		n.Subs = make(map[string]*Node)
	}
	subName := subPath[0]
	subNode, ok := n.Subs[subName]
	if !ok {
		subNode = NewTree(subPath[1:], t)
		n.Subs[subName] = subNode
	} else {
		subNode.Update(subPath[1:], t)
	}
}

func (n *Node) Get(subPath []string) (time.Time, bool) {
	if len(subPath) == 0 {
		return n.ModTime, true
	}
	subName := subPath[0]
	subNode, ok := n.Subs[subName]
	if !ok {
		return time.Time{}, false
	}
	return subNode.Get(subPath[1:])
}

func (n *Node) Delete(subPath []string) {
	if len(subPath) == 0 {
		panic("impossible")
	}
	subName := subPath[0]
	if len(subPath) == 1 {
		// leaf
		delete(n.Subs, subName)
	} else {
		subNode, ok := n.Subs[subName]
		if !ok {
			return
		}
		subNode.Delete(subPath[1:])
	}
}

func (n *Node) dump(name string, level int) {
	pt("%s%s %v\n", strings.Repeat("  ", level), name, n.ModTime)
	for name, sub := range n.Subs {
		sub.dump(name, level+1)
	}
}

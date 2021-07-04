// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package ling

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"reflect"
	"runtime"
	"runtime/pprof"
	"strings"
	"testing"

	"github.com/reusee/dscope"
	"github.com/reusee/e4"
	"github.com/reusee/ling/v2/index"
	"github.com/reusee/ling/v2/store"
	"github.com/reusee/ling/v2/storedisk"
	"github.com/reusee/ling/v2/storekv"
	"github.com/reusee/ling/v2/storemem"
	"github.com/reusee/ling/v2/storepebble"
	"github.com/reusee/ling/v2/sys"
	"github.com/reusee/ling/v2/vars"
	"github.com/reusee/pr"
)

func init() {
	go func() {
		ce(http.ListenAndServe("127.0.0.1:9991", nil))
	}()
}

func TestMain(m *testing.M) {

	runtime.SetBlockProfileRate(10 * 1000)
	runtime.MemProfileRate = 64 * 1024

	ret := m.Run()

	for _, profile := range []string{
		"block",
		"heap",
		"allocs",
	} {
		w, err := os.Create("profile." + profile)
		ce(err)
		p := pprof.Lookup(profile)
		if p == nil {
			panic(fmt.Errorf("no such profile: %s", profile))
		}
		ce(p.WriteTo(w, 0))
		ce(w.Close())
	}

	os.Exit(ret)
}

var (
	storeType = reflect.TypeOf((*store.Store)(nil)).Elem()
)

func runTest(
	t *testing.T,
	fn any,
) {
	t.Helper()

	// test func
	fnValue := reflect.ValueOf(fn)
	fnType := fnValue.Type()

	// store and index
	var stores []any
	for i := 0; i < fnType.NumIn(); i++ {
		argType := fnType.In(i)
		if argType == storeType {
			stores = storeDefs
		}
	}
	if len(stores) == 0 {
		stores = []any{memStore}
	}

	// specs
	type Spec struct {
		Desc []string
		Defs []any
	}
	var specs []Spec
	for i, store := range stores {
		for j, manager := range indexManagerDefs {
			defs := append(Defs[:0:0], Defs...)
			defs = append(defs, store, manager)
			specs = append(specs, Spec{
				Desc: []string{
					sp("store-%d", i),
					sp("index-%d", j),
				},
				Defs: defs,
			})
		}
	}

	// run
	if len(specs) == 1 {
		spec := specs[0]
		waitTree := pr.NewWaitTree(nil, nil)
		spec.Defs = append(spec.Defs,
			func() *testing.T {
				return t
			},
			func() vars.VarsDir {
				return vars.VarsDir(t.TempDir())
			},
			func() (context.Context, *pr.WaitTree) {
				return waitTree.Ctx, waitTree
			},
		)
		scope := dscope.New(spec.Defs...).Sub(
			func() sys.Testing {
				return true
			},
		)
		scope.Call(fn)
		waitTree.Cancel()
		waitTree.Wait()

	} else {
		for _, spec := range specs {
			spec := spec
			t.Run(
				strings.Join(spec.Desc, ":"),
				func(t *testing.T) {
					t.Parallel()
					waitTree := pr.NewWaitTree(nil, nil)
					spec.Defs = append(spec.Defs,
						func() *testing.T {
							return t
						},
						func() vars.VarsDir {
							return vars.VarsDir(t.TempDir())
						},
						func() (context.Context, *pr.WaitTree) {
							return waitTree.Ctx, waitTree
						},
					)
					scope := dscope.New(spec.Defs...).Sub(
						func() sys.Testing {
							return true
						},
					)
					scope.Call(fn)
					waitTree.Cancel()
					waitTree.Wait()
				},
			)
		}
	}

}

var indexManagerDefs = []any{
	// mem
	0: memIndexManager,

	// mem pebble
	1: func(
		t *testing.T,
		newPebble storepebble.New,
		wt *pr.WaitTree,
	) index.IndexManager {
		defer he(nil, e4.TestingFatal(t))
		peb, err := newPebble(storepebble.NewMemFS(), "foo")
		ce(err)
		done := wt.Add()
		go func() {
			defer done()
			<-wt.Ctx.Done()
			ce(peb.Close())
		}()
		return peb
	},

	// mem pebble batch
	2: func(
		t *testing.T,
		newPebble storepebble.New,
		wt *pr.WaitTree,
		newBatch storepebble.NewBatch,
	) index.IndexManager {
		defer he(nil, e4.TestingFatal(t))
		peb, err := newPebble(storepebble.NewMemFS(), "foo")
		ce(err)
		batch, err := newBatch(peb)
		ce(err)
		done := wt.Add()
		go func() {
			defer done()
			<-wt.Ctx.Done()
			ce(batch.Close())
			ce(peb.Close())
		}()
		return batch
	},
}

var memIndexManager = func(
	newMemStore storemem.New,
) index.IndexManager {
	return newMemStore()
}

var storeDefs = []any{
	// disk pebble
	0: func(
		t *testing.T,
		newPebble storepebble.New,
		newKV storekv.New,
		wt *pr.WaitTree,
	) store.Store {
		defer he(nil, e4.TestingFatal(t))
		dir := t.TempDir()
		peb, err := newPebble(nil, dir)
		ce(err)
		done := wt.Add()
		go func() {
			defer done()
			<-wt.Ctx.Done()
			ce(peb.Close())
		}()
		s, err := newKV(peb, "foo")
		ce(err)
		return s
	},

	// disk
	1: func(
		t *testing.T,
		newDiskStore storedisk.New,
		newKV storekv.New,
		wt *pr.WaitTree,
	) store.Store {
		defer he(nil, e4.TestingFatal(t))
		dir := t.TempDir()
		s, err := newDiskStore(dir)
		ce(err)
		done := wt.Add()
		go func() {
			defer done()
			<-wt.Ctx.Done()
			ce(s.Close())
		}()
		kv, err := newKV(s, "foo")
		ce(err)
		return kv
	},

	// mem
	2: memStore,

	// disk with cache
	3: func(
		t *testing.T,
		newDiskStore storedisk.New,
		newKV storekv.New,
		newMemCache store.NewMemCache,
		wt *pr.WaitTree,
	) store.Store {
		defer he(nil, e4.TestingFatal(t))
		dir := t.TempDir()
		s, err := newDiskStore(dir)
		ce(err)
		done := wt.Add()
		go func() {
			defer done()
			<-wt.Ctx.Done()
			ce(s.Close())
		}()
		cache, err := newMemCache(1024, 2048)
		ce(err)
		kv, err := newKV(
			s,
			"foo",
			storekv.WithCache(cache),
		)
		ce(err)
		return kv
	},

	// pebble batch
	4: func(
		t *testing.T,
		newPebble storepebble.New,
		wt *pr.WaitTree,
		newBatch storepebble.NewBatch,
		newKV storekv.New,
	) store.Store {
		defer he(nil, e4.TestingFatal(t))
		peb, err := newPebble(storepebble.NewMemFS(), "foo")
		ce(err)
		batch, err := newBatch(peb)
		ce(err)
		kv, err := newKV(batch, "foo")
		ce(err)
		done := wt.Add()
		go func() {
			defer done()
			<-wt.Ctx.Done()
			ce(kv.Close())
			ce(batch.Close())
			ce(peb.Close())
		}()
		return kv
	},
}

var memStore = func(
	newMem storemem.New,
	newKV storekv.New,
) store.Store {
	s, err := newKV(newMem(), "foo")
	ce(err)
	return s
}

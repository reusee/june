// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package june

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
	"github.com/reusee/e5"
	"github.com/reusee/june/index"
	"github.com/reusee/june/store"
	"github.com/reusee/june/storedisk"
	"github.com/reusee/june/storekv"
	"github.com/reusee/june/storemem"
	"github.com/reusee/june/storepebble"
	"github.com/reusee/june/sys"
	"github.com/reusee/june/vars"
	"github.com/reusee/pr2"
)

func init() {
	go func() {
		ce(http.ListenAndServe("127.0.0.1:65400", nil))
	}()
}

func TestMain(m *testing.M) {

	runtime.SetBlockProfileRate(10 * 1000)
	runtime.SetMutexProfileFraction(1)
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
		wg := pr2.NewWaitGroup(context.Background())
		spec.Defs = append(spec.Defs,
			func() *testing.T {
				return t
			},
			func() vars.VarsSpec {
				return func() (string, *pr2.WaitGroup) {
					return t.TempDir(), wg
				}
			},
			func() *pr2.WaitGroup {
				return wg
			},
		)
		scope := dscope.New(spec.Defs...).Fork(
			func() sys.Testing {
				return true
			},
		)
		scope.Call(fn)
		wg.Cancel()
		wg.Wait()

	} else {
		for _, spec := range specs {
			spec := spec
			t.Run(
				strings.Join(spec.Desc, ":"),
				func(t *testing.T) {
					t.Parallel()
					wg := pr2.NewWaitGroup(context.Background())
					spec.Defs = append(spec.Defs,
						func() *testing.T {
							return t
						},
						func() vars.VarsSpec {
							return func() (string, *pr2.WaitGroup) {
								return t.TempDir(), wg
							}
						},
						func() *pr2.WaitGroup {
							return wg
						},
					)
					scope := dscope.New(spec.Defs...).Fork(
						func() sys.Testing {
							return true
						},
					)
					scope.Call(fn)
					wg.Cancel()
					wg.Wait()
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
		wg *pr2.WaitGroup,
	) index.IndexManager {
		defer he(nil, e5.TestingFatal(t))
		peb, err := newPebble(wg, storepebble.NewMemFS(), "foo")
		ce(err)
		return peb
	},

	// mem pebble batch
	2: func(
		t *testing.T,
		newPebble storepebble.New,
		wg *pr2.WaitGroup,
		newBatch storepebble.NewBatch,
	) index.IndexManager {
		defer he(nil, e5.TestingFatal(t))
		ctx := pr2.NewWaitGroup(context.TODO())
		peb, err := newPebble(wg, storepebble.NewMemFS(), "foo")
		ce(err)
		batch, err := newBatch(ctx, peb)
		ce(err)
		return batch
	},
}

var memIndexManager = func(
	newMemStore storemem.New,
) index.IndexManager {
	ctx := pr2.NewWaitGroup(context.TODO())
	return newMemStore(ctx)
}

var storeDefs = []any{
	// mem pebble
	0: func(
		t *testing.T,
		newPebble storepebble.New,
		newKV storekv.New,
		wg *pr2.WaitGroup,
	) store.Store {
		defer he(nil, e5.TestingFatal(t))
		peb, err := newPebble(wg, storepebble.NewMemFS(), "peb")
		ce(err)
		s, err := newKV(wg, peb, "foo")
		ce(err)
		return s
	},

	// disk
	1: func(
		t *testing.T,
		newDiskStore storedisk.New,
		newKV storekv.New,
		wg *pr2.WaitGroup,
	) store.Store {
		defer he(nil, e5.TestingFatal(t))
		dir := t.TempDir()
		s, err := newDiskStore(wg, dir)
		ce(err)
		kv, err := newKV(wg, s, "foo")
		ce(err)
		return kv
	},

	// mem
	2: memStore,

	// pebble batch
	3: func(
		t *testing.T,
		newPebble storepebble.New,
		newBatch storepebble.NewBatch,
		newKV storekv.New,
		wg *pr2.WaitGroup,
	) store.Store {
		defer he(nil, e5.TestingFatal(t))
		peb, err := newPebble(wg, storepebble.NewMemFS(), "foo")
		ce(err)
		batch, err := newBatch(wg, peb)
		ce(err)
		kv, err := newKV(wg, batch, "foo")
		ce(err)
		return kv
	},
}

var memStore = func(
	newMem storemem.New,
	newKV storekv.New,
	wg *pr2.WaitGroup,
) store.Store {
	s, err := newKV(wg, newMem(wg), "foo")
	ce(err)
	return s
}

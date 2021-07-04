// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package config

import (
	"reflect"

	"github.com/reusee/e4"
	"github.com/reusee/starlarkutil"
	"go.starlark.net/resolve"
	"go.starlark.net/starlark"
)

func init() {
	resolve.AllowNestedDef = true
	resolve.AllowLambda = true
	resolve.AllowFloat = true
	resolve.AllowSet = true
	resolve.AllowGlobalReassign = true
	resolve.AllowRecursion = true
}

type UserConfig []byte

func (_ Def) UserConfig() UserConfig {
	return []byte{}
}

type GetConfig func(target interface{}) error

func (_ Def) Get(
	userConfig UserConfig,
) (
	get GetConfig,
) {

	exec := func(src []byte) (_ starlark.StringDict, err error) {
		defer he(&err)
		globals, err := starlark.ExecFile(new(starlark.Thread), "config.star", src, starlark.StringDict{
			"foo": starlark.String("foo"),
		})
		ce(err)
		return globals, nil
	}

	userGlobals, err := exec(userConfig)
	ce(err)

	defaultGlobals, err := exec(DefaultConfig)
	ce(err)

	get = func(target interface{}) (err error) {
		defer he(&err)

		defer catch(&err,
			e4.NewInfo("type: %v", reflect.TypeOf(target).Elem()),
		)

		value := reflect.ValueOf(target)
		t := value.Type()
		if t.Kind() != reflect.Ptr {
			panic("must be pointer")
		}
		if t.Elem().Kind() != reflect.Struct {
			panic("must be struct")
		}
		value = value.Elem()
		t = t.Elem()

		for i := 0; i < value.NumField(); i++ {
			name := t.Field(i).Name
			v, ok := userGlobals[name]
			if !ok {
				v, ok = defaultGlobals[name]
				if !ok {
					return
				}
			}
			if v == nil {
				return
			}
			err := starlarkutil.Assign(v, value.Field(i).Addr())
			ce(err)
		}

		return nil
	}

	return
}

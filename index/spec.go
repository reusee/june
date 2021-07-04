// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package index

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/reusee/june/naming"
	"github.com/reusee/sb"
)

type Spec struct {
	Name   string
	Type   reflect.Type
	Fields []reflect.Type
}

var specsByType sync.Map

var specsByName sync.Map

var sbMarshalerType = reflect.TypeOf((*sb.SBMarshaler)(nil)).Elem()

func RegisterIndex(name string, t reflect.Type) {
	spec := Spec{
		Type: t,
		Name: name,
	}

	switch t.Kind() {

	case reflect.Struct:
		numFields := t.NumField()
		for i := 0; i < numFields; i++ {
			spec.Fields = append(spec.Fields, t.Field(i).Type)
		}

	case reflect.Func:
		numIn := t.NumIn()
		for i := 0; i < numIn; i++ {
			spec.Fields = append(spec.Fields, t.In(i))
		}

	default:
		panic(fmt.Errorf("index should be struct or function kind"))
	}

	specsByType.Store(t, spec)
	specsByName.Store(name, spec)
}

func Register(o any) {
	t := reflect.TypeOf(o)

	name := naming.Type(t)

	if name == "" {
		panic(fmt.Errorf("index should be defined type"))
	}

	RegisterIndex(name, t)

}

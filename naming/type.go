// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package naming

import (
	"fmt"
	"reflect"
	"runtime/debug"
	"strings"
	"sync"
)

var (
	typeToName sync.Map
	nameToType sync.Map
)

//TODO return migrated names
func Type(t reflect.Type) (name string) {
	if v, ok := typeToName.Load(t); ok {
		return v.(string)
	}

	defer func() {
		if name != "" {
			typeToName.Store(t, name)
			nameToType.Store(name, t)
		}
	}()

	// pointer
	if t.Kind() == reflect.Ptr {
		return "*" + Type(t.Elem())
	}

	// defined types
	if definedName := t.Name(); definedName != "" {
		if pkgPath := t.PkgPath(); pkgPath != "" {
			if buildPackagePath != "" && pkgPath == buildPackagePath {
				// when testing a main package, pkgPath will be the real path instead of 'main'
				// replace with 'main' to ensure compatibility with existed data
				pkgPath = "main"
			}
			pkgPath = strings.TrimPrefix(pkgPath, DefaultPath)
			pkgPath = strings.TrimPrefix(pkgPath, DefaultDomain)
			return pkgPath + "." + definedName
		} else {
			return definedName
		}
	}

	panic(fmt.Errorf("not defined type: %s", t.String()))

}

var buildPackagePath = func() string {
	info, _ := debug.ReadBuildInfo()
	if info == nil {
		return ""
	}
	return info.Path
}()

func GetType(name string) *reflect.Type {
	v, ok := nameToType.Load(name)
	if !ok {
		return nil
	}
	t := v.(reflect.Type)
	return &t
}

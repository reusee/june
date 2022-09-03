// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

//go:build ignore
// +build ignore

package main

import (
	"bytes"
	"fmt"
	"go/format"
	"go/types"
	"os"
	"sort"
	"strings"

	"github.com/reusee/e5"
	"golang.org/x/tools/go/packages"
)

var (
	pt     = fmt.Printf
	ce, he = e5.CheckWithStacktrace, e5.Handle
)

func main() {
	pkgs, err := packages.Load(
		&packages.Config{
			Mode: packages.NeedTypesInfo |
				packages.NeedTypes |
				packages.NeedFiles |
				packages.NeedSyntax |
				packages.NeedName,
			BuildFlags: []string{
				"-tags=step2",
			},
		},
		"./...",
		"testing",
	)
	ce(err)
	if packages.PrintErrors(pkgs) > 0 {
		return
	}

	var testingTType types.Type
	packages.Visit(pkgs, func(pkg *packages.Package) bool {
		if pkg.Name != "testing" {
			return true
		}
		testingTType = types.NewPointer(pkg.Types.Scope().Lookup("T").Type())
		return false
	}, nil)
	if testingTType == nil {
		panic("testing.T not found")
	}

	type Func struct {
		Name    string
		PkgName string
		PkgPath string
		OS      string
	}

	funcsByOS := make(map[string][]Func)
	packages.Visit(pkgs, func(pkg *packages.Package) bool {
		if pkg.Name == "testing" {
			return true
		}
		globalScope := pkg.Types.Scope()
		for _, name := range globalScope.Names() {
			obj := globalScope.Lookup(name)
			fn, ok := obj.(*types.Func)
			if !ok {
				continue
			}
			if !strings.HasPrefix(fn.Name(), "Test") {
				continue
			}
			signature := fn.Type().(*types.Signature)
			params := signature.Params()
			if params.Len() < 1 {
				continue
			}
			p1 := params.At(0)
			if !types.Identical(p1.Type(), testingTType) {
				continue
			}

			var os string
			pos := fn.Pos()
			file := pkg.Fset.File(pos)
			filename := file.Name()
			if strings.HasSuffix(filename, "_windows.go") {
				os = "windows"
			}

			funcsByOS[os] = append(
				funcsByOS[os],
				Func{
					Name:    fn.Name(),
					PkgName: pkg.Name,
					PkgPath: pkg.PkgPath,
					OS:      os,
				},
			)
		}
		return true
	}, nil)

	for OS, fns := range funcsByOS {
		sort.Slice(fns, func(i, j int) bool {
			f1 := fns[i]
			f2 := fns[j]
			if f1.PkgPath != f2.PkgPath {
				return f1.PkgPath < f2.PkgPath
			}
			return f1.Name < f2.Name
		})

		imports := make(map[string]bool)
		for _, fn := range fns {
			imports[fn.PkgPath] = true
		}

		buf := new(bytes.Buffer)
		_, err = buf.WriteString(`// +build !step1 !step2 ` + OS + `

		package june

		import (
			` + func() string {
			buf := new(strings.Builder)
			for imp := range imports {
				buf.WriteString(`"` + imp + `"` + "\n")
			}
			return buf.String()
		}() + `
			"testing"
		)
		`)
		ce(err)

		for _, fn := range fns {
			_, err = buf.WriteString(`
		func Test_` + fn.PkgName + `_` + fn.Name + `(t *testing.T) {
		  t.Parallel()
		  runTest(t, ` + fn.PkgName + "." + fn.Name + `)
		}
		    `)
			ce(err)
		}

		src, err := format.Source(buf.Bytes())
		ce(err)
		var outFileName string
		if OS != "" {
			outFileName = "all_" + OS + "_test.go"
		} else {
			outFileName = "all_test.go"
		}
		ce(os.WriteFile(outFileName, src, 0644))

	}

}

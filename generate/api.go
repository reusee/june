// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

//go:build ignore
// +build ignore

package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/types"
	"os"
	"sort"
	"strings"

	"github.com/reusee/e4"
	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
)

var (
	pt = fmt.Printf
	ce = e4.Check
)

//TODO generate Is*Option docs
//TODO skip non-exported methods
//TODO merge reducers

func main() {
	pkgs, err := packages.Load(
		&packages.Config{
			Mode: packages.NeedTypesInfo |
				packages.NeedFiles |
				packages.NeedSyntax |
				packages.NeedTypes |
				packages.NeedName,
			BuildFlags: []string{
				"-tags=step1",
			},
		},
		"./...",
	)
	ce(err)
	if packages.PrintErrors(pkgs) > 0 {
		return
	}

	var provides []string
	packages.Visit(pkgs, func(pkg *packages.Package) bool {
		defObject := pkg.Types.Scope().Lookup("Def")
		if defObject == nil {
			return true
		}
		defMethodSet := types.NewMethodSet(defObject.Type())

		for i := 0; i < defMethodSet.Len(); i++ {
			sel := defMethodSet.At(i)
			signature := sel.Type().(*types.Signature)

			// doc
			rets := signature.Results()
			for i := 0; i < rets.Len(); i++ {
				retType := rets.At(i).Type()
				// header
				provide := types.TypeString(retType, func(pkg *types.Package) string {
					return pkg.Name()
				})
				// type
				provide += "\n\t" + types.TypeString(retType.Underlying(), func(pkg *types.Package) string {
					return pkg.Name()
				})
				// doc
				for id, obj := range pkg.TypesInfo.Defs {
					if obj == nil {
						continue
					}
					typeName, ok := obj.(*types.TypeName)
					if !ok {
						continue
					}
					if types.Identical(retType, typeName.Type()) {
						for _, file := range pkg.Syntax {
							path, _ := astutil.PathEnclosingInterval(file, id.Pos(), id.Pos())
							for _, node := range path {
								spec, ok := node.(*ast.TypeSpec)
								if ok && spec.Doc != nil {
									provide += "\n\t" + strings.TrimSpace(spec.Doc.Text())
								}
								decl, ok := node.(*ast.GenDecl)
								if ok && decl.Doc != nil {
									provide += "\n\t" + strings.TrimSpace(decl.Doc.Text())
								}
							}
						}
					}
				}
				provides = append(provides, provide)
			}

		}
		return true
	}, nil)

	sort.Slice(provides, func(i, j int) bool {
		return provides[i] < provides[j]
	})
	buf := new(bytes.Buffer)
	for _, provide := range provides {
		buf.WriteString(provide)
		buf.WriteString("\n")
		buf.WriteString("\n")
	}
	err = os.WriteFile("api", buf.Bytes(), 0644)
	ce(err)

}

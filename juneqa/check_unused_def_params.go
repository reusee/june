// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package juneqa

import (
	"fmt"
	"go/ast"
	"go/types"

	"github.com/reusee/qa"
	"golang.org/x/tools/go/packages"
)

func (_ Def) CheckUnusedDefParams(
	pkgs []*packages.Package,
) qa.CheckFunc {
	return func() (ret []error) {

		// sys.Def type
		var sysDefType types.Type
		for _, pkg := range pkgs {
			if pkg.Name != "sys" {
				continue
			}
			sysDefType = pkg.Types.Scope().Lookup("Def").(*types.TypeName).Type()
		}
		if sysDefType == nil {
			panic("sys.Def not found")
		}

		// find unused injected params
		for _, pkg := range pkgs {
			for _, file := range pkg.Syntax {
				for _, decl := range file.Decls {

					fnDecl, ok := decl.(*ast.FuncDecl)
					if !ok {
						continue
					}
					obj := pkg.TypesInfo.Defs[fnDecl.Name]
					sig := obj.Type().(*types.Signature)
					params := sig.Params()
					if params.Len() == 0 {
						continue
					}
					if params.At(0).Type() != sysDefType {
						continue
					}

					args := fnDecl.Type.Params.List
					type Obj struct {
						Ident *ast.Ident
						Obj   types.Object
					}
					objs := make(map[types.Object]Obj)
					for _, arg := range args[1:] {
						for _, ident := range arg.Names {
							obj := pkg.TypesInfo.Defs[ident]
							objs[obj] = Obj{
								Ident: ident,
								Obj:   obj,
							}
						}
					}

					ast.Inspect(fnDecl.Body, func(node ast.Node) bool {
						ident, ok := node.(*ast.Ident)
						if !ok {
							return true
						}
						obj := pkg.TypesInfo.Uses[ident]
						if obj == nil {
							return true
						}
						delete(objs, obj)
						return true
					})

					for _, obj := range objs {
						if obj.Ident.Name == "_" {
							continue
						}
						pos := pkg.Fset.Position(obj.Ident.Pos())
						ret = append(ret, fmt.Errorf(
							"unused param: %s at %s:%d",
							obj.Ident.Name,
							pos.Filename,
							pos.Line,
						))
					}

				}
			}
		}

		return
	}
}

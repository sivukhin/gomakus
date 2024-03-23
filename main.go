package main

import (
	"fmt"
	"go/ast"
	"log"

	"golang.org/x/tools/go/packages"

	"gomakus/src"
)

func main() {
	cfg := &packages.Config{
		Mode:  packages.NeedSyntax | packages.NeedFiles | packages.NeedTypes,
		Tests: false,
		Dir:   ".",
	}

	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		panic(fmt.Errorf("failed to load package: %v", cfg))
	}

	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			ast.Inspect(file, func(node ast.Node) bool {
				funcDecl, ok := node.(*ast.FuncDecl)
				if !ok {
					return true
				}
				if funcDecl.Body == nil {
					return true
				}
				execution := src.ExecutionFromFunc(src.NewScopes(map[string]src.FuncId{
					src.SliceFuncName:  src.SliceFuncId,
					src.AppendFuncName: src.AppendFuncId,
				}), pkg.Fset, funcDecl)
				warnings := src.ValidateExecution(src.DefaultFuncSpecCollection, execution)
				for _, warning := range warnings {
					pos, ok := execution.SourceCodeReferences.References[warning.ExecutionPoint]
					if ok {
						log.Printf(
							"potential append overwrite found in function: func=[%v], file=[%v], line=[%v]",
							funcDecl.Name,
							pkg.Fset.Position(pos).Filename,
							pkg.Fset.Position(pos).Line,
						)
					} else {
						log.Printf(
							"potential append overwrite found: func=[%v], file=[%v]",
							funcDecl.Name,
							pkg.Fset.Position(funcDecl.Pos()).Filename,
						)
					}
				}
				return true
			})
		}
	}
}

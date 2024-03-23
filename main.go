package main

import (
	"flag"
	"fmt"
	"go/ast"
	"log"
	"os"
	"path/filepath"

	"golang.org/x/tools/go/packages"

	"github.com/sivukhin/gomakus/src"
)

func reportWarning(format string, analysisPath string, fileName, funcName string, line int) {
	if format == "github" {
		relativePath, _ := filepath.Rel(analysisPath, fileName)
		fmt.Printf("::warning file=%v,line=%v::%v\n", relativePath, line, "potential append overwrite found")
	} else {
		log.Printf(
			"potential append overwrite found: func=[%v], file=[%v], line=[%v]",
			funcName,
			fileName,
			line,
		)
	}
}

func main() {
	modulePath := flag.String("path", "", "path to the module root (with go.mod file)")
	reportFormat := flag.String("format", "log", "reporting type (github | log)")
	flag.Parse()

	var analysisPath string
	var err error
	if *modulePath == "" {
		analysisPath, err = os.Getwd()
		if err != nil {
			fmt.Printf("unable to get working directory: %v\n", err)
			flag.Usage()
			os.Exit(1)
		}
	} else {
		analysisPath, err = filepath.Abs(*modulePath)
		if err != nil {
			fmt.Printf("unable to expand path '%v' to absolute: %v\n", *modulePath, err)
			flag.Usage()
			os.Exit(1)
		}
	}

	cfg := &packages.Config{
		Mode:  packages.NeedSyntax | packages.NeedFiles | packages.NeedTypes,
		Tests: false,
		Dir:   analysisPath,
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
				_ = execution
				warnings := src.ValidateExecution(src.DefaultFuncSpecCollection, execution)
				for _, warning := range warnings {
					pos, ok := execution.SourceCodeReferences.References[warning.ExecutionPoint]
					if ok {
						reportWarning(*reportFormat, analysisPath, pkg.Fset.Position(pos).Filename, funcDecl.Name.Name, pkg.Fset.Position(pos).Line)
					} else {
						reportWarning(*reportFormat, analysisPath, pkg.Fset.Position(pos).Filename, funcDecl.Name.Name, pkg.Fset.Position(funcDecl.Pos()).Line)
					}
				}
				return true
			})
		}
	}
}

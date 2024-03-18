package utils

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

func MustGenStatements(statements string) (*token.FileSet, []ast.Stmt) {
	fset, funcAst := MustGenFunc(fmt.Sprintf("func main() {\n%v\n}", statements))
	return fset, funcAst.Body.List
}

func MustGenFunc(src string) (*token.FileSet, *ast.FuncDecl) {
	fset, fileAst := MustGenSrc(fmt.Sprintf("package main\n%v", src))
	return fset, fileAst.Decls[0].(*ast.FuncDecl)
}

func MustGenSrc(src string) (*token.FileSet, *ast.File) {
	fset := token.NewFileSet()
	fileAst, err := parser.ParseFile(fset, "", src, parser.AllErrors)
	if err != nil {
		panic(fmt.Errorf("src parsing failed: %w", err))
	}
	return fset, fileAst
}

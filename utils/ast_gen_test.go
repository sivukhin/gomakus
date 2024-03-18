package utils

import (
	"testing"
)

func TestGenStatements(t *testing.T) {
	fset, statements := MustGenStatements(`a := 1
b := make([]int, 3, 10)
c := a + len(b)
if a == b {
panic("a == b")
}
`)
	t.Log(fset, statements)
}

func TestGenFunc(t *testing.T) {
	fset, funcAst := MustGenFunc(`func sum(a, b int) int {
return a + b
}`)
	t.Log(fset, funcAst)
}

func TestGenSrc(t *testing.T) {
	fset, srcAst := MustGenSrc(`package main
func sub(a, b int) int {
return a - b
}
func sum(a, b int) int {
return a + b
}`)
	t.Log(fset, srcAst)
}

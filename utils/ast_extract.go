package utils

import (
	"fmt"
	"go/ast"
)

func MustExtractLabeledStatement[T ast.Node](label string, nodes ...T) ast.Node {
	var labeled ast.Node
	for _, node := range nodes {
		ast.Inspect(node, func(node ast.Node) bool {
			if label, ok := node.(*ast.LabeledStmt); ok {
				labeled = label.Stmt
			}
			return true
		})
	}
	if labeled == nil {
		panic(fmt.Errorf("unable to find labeled statment '%v'", label))
	}
	return labeled
}

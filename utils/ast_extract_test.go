package utils

import (
	"go/ast"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLabeledExtraction(t *testing.T) {
	_, statements := MustGenStatements(`label: if 1 == 1 {
panic("1 == 1")
} else if 2 == 2 {
panic("2 == 2")
} else {
panic("no panic")
}`)
	statement := MustExtractLabeledStatement("label", statements...)
	require.IsType(t, &ast.IfStmt{}, statement)
	t.Log(statement)
}

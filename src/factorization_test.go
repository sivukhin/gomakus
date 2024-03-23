package src

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sivukhin/gomakus/utils"
)

func TestFactorizeAssignments(t *testing.T) {
	// b = a.x
	// c = b
	// d = c.y
	// e = a.y
	rules := FactorizeAssignments([]AssignSelectorOp{
		{
			FromSelector: VarSelector{VarId: 0, Selector: Path{"x"}},
			ToSelector:   VarSelector{VarId: 1},
		},
		{
			FromSelector: VarSelector{VarId: 1},
			ToSelector:   VarSelector{VarId: 2},
		},
		{
			FromSelector: VarSelector{VarId: 2, Selector: Path{"y"}},
			ToSelector:   VarSelector{VarId: 3},
		},
		{
			FromSelector: VarSelector{VarId: 0, Selector: Path{"y"}},
			ToSelector:   VarSelector{VarId: 4},
		},
	})
	t.Log(rules)
	require.Equal(t, FactorizationRules{
		0: {{"x", "y"}, {"y"}},
		1: {{"y"}},
		2: {{"y"}},
		3: {{}},
		4: {{}},
	}, rules)
}

func TestFactorizationHard(t *testing.T) {
	fset, funcDecl := utils.MustGenFunc(`func executionFromExpr(
	builder ExecutionBuilder,
	scopes Scopes,
	fset *token.FileSet,
	expr ast.Expr,
	exprOutputs int,
) (ExecutionBuilder, []VarComposition) {
	blanks := make([]VarComposition, exprOutputs)

	switch e := expr.(type) {
	case *ast.CallExpr:
		var args []ast.Expr
		call := e.(*ast.CallExpr)
		args = call.Args
	}
	panic(fmt.Errorf("unexpected expression"))
}
`)
	execution := ExecutionFromFunc(NewScopes(map[string]FuncId{
		SliceFuncName:  SliceFuncId,
		AppendFuncName: AppendFuncId,
	}), fset, funcDecl)
	t.Logf("%v", execution)
	assigns := SelectAssignOps(SimplificationContext{Funcs: DefaultFuncSpecCollection}, execution)
	t.Logf("%+v", assigns)
	factorization := FactorizeAssignments(assigns)
	t.Log(factorization)
}

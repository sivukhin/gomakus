package src

import (
	"testing"

	"github.com/stretchr/testify/require"

	"gomakus/utils"
)

func TestExecutionTrace(t *testing.T) {
	fset, funcDecl := utils.MustGenFunc(`func parseUnnestedKeyFieldSet(raw string, prefix []string) Set {
	ret := Set{}

	for _, s := range strings.Fields(raw) {
		next := append(prefix[:], s)
		ret = append(ret, next)
	}
	return ret
}`)
	execution := ExecutionFromFunc(NewScopes(map[string]FuncId{
		SliceFuncName:  SliceFuncId,
		AppendFuncName: AppendFuncId,
	}), fset, funcDecl)
	t.Logf("%v", execution)
	require.NotEmpty(t, ValidateExecution(DefaultFuncSpecCollection, execution))
}

func TestTraceValidation(t *testing.T) {
	require.Empty(t, ValidateTrace(ExecutionTrace{
		{ToPoint: 1, Operation: AssignVarOp{FromVarId: BlankVarId, ToVarId: 0, GenChange: 0}},
		{ToPoint: 2, Operation: AssignVarOp{FromVarId: BlankVarId, ToVarId: 1, GenChange: 0}},
		{ToPoint: 3, Operation: AssignVarOp{FromVarId: 0, ToVarId: 2, GenChange: 1}},
		{ToPoint: 4, Operation: AssignVarOp{FromVarId: 2, ToVarId: 0, GenChange: 0}},
		{ToPoint: 5, Operation: AssignVarOp{FromVarId: 1, ToVarId: 3, GenChange: 1}},
		{ToPoint: 6, Operation: AssignVarOp{FromVarId: 3, ToVarId: 1, GenChange: 0}},
		{ToPoint: 7, Operation: AssignVarOp{FromVarId: 0, ToVarId: 2, GenChange: 1}},
		{ToPoint: 8, Operation: AssignVarOp{FromVarId: 2, ToVarId: 0, GenChange: 0}},
		{ToPoint: 9, Operation: AssignVarOp{FromVarId: 1, ToVarId: 3, GenChange: 1}},
		{ToPoint: 10, Operation: AssignVarOp{FromVarId: 3, ToVarId: 1, GenChange: 0}},
	}))
}

func TestExecutionTrace2(t *testing.T) {
	fset, funcDecl := utils.MustGenFunc(`func deconstructIf(fset *token.FileSet, ifStmt *ast.IfStmt) (inits []ast.Stmt, bodies []ast.Stmt) {
	inits, bodies = make([]ast.Stmt, 0), make([]ast.Stmt, 0)
	for ifStmt != nil {
		inits = append(inits, ifStmt.Init)
		bodies = append(bodies, ifStmt.Body)
	}
	return
}
`)
	execution := ExecutionFromFunc(NewScopes(map[string]FuncId{
		SliceFuncName:  SliceFuncId,
		AppendFuncName: AppendFuncId,
	}), fset, funcDecl)
	t.Logf("%v", execution)
	warnings := ValidateExecution(DefaultFuncSpecCollection, execution)
	require.Empty(t, warnings)
}
func TestExecutionTrace4(t *testing.T) {
	fset, funcDecl := utils.MustGenFunc(`func factorizeAssigment(context factorizationContext, varId VarId, path Path) {
	for _, source := range context.sources[varId] {
		targetPath := append(append([]string{}, source.FromSelector.Selector...), path[len(source.ToSelector.Selector):]...)
	}
}
`)
	execution := ExecutionFromFunc(NewScopes(map[string]FuncId{
		SliceFuncName:  SliceFuncId,
		AppendFuncName: AppendFuncId,
	}), fset, funcDecl)
	t.Logf("%v", execution)
	warnings := ValidateExecution(DefaultFuncSpecCollection, execution)
	require.Empty(t, warnings)
}

func TestExecutionTrace3(t *testing.T) {
	fset, funcDecl := utils.MustGenFunc(`func factorizeAssigment(context factorizationContext, varId VarId, path Path) {
	pathBytes := joinTo(context.workspace, path, ",")
	for _, source := range context.sources[varId] {
		targetPath := append(append([]string{}, source.FromSelector.Selector...), path[len(source.ToSelector.Selector):]...)
	}
}
`)
	execution := ExecutionFromFunc(NewScopes(map[string]FuncId{
		SliceFuncName:  SliceFuncId,
		AppendFuncName: AppendFuncId,
	}), fset, funcDecl)
	t.Logf("%v", execution)
	warnings := ValidateExecution(DefaultFuncSpecCollection, execution)
	require.Empty(t, warnings)
}

func TestExecutionTrace5(t *testing.T) {
	fset, funcDecl := utils.MustGenFunc(`func SelectAssignOps(context SimplificationContext, execution Execution) []AssignSelectorOp {
	assigns := make([]AssignSelectorOp, 0)
	switch op := transition.Operation.(type) {
	case AssignSelectorOp:
		assigns = append(assigns, op)
	case UseSelectorsOp:
		assigns = append(assigns, AssignSelectorOp{})
	}
	return assigns
}
`)
	execution := ExecutionFromFunc(NewScopes(map[string]FuncId{
		SliceFuncName:  SliceFuncId,
		AppendFuncName: AppendFuncId,
	}), fset, funcDecl)
	t.Logf("%v", execution)
	warnings := ValidateExecution(DefaultFuncSpecCollection, execution)
	require.Empty(t, warnings)
}

func TestKek(t *testing.T) {
	/*
		2:src.AssignVarOp{FromVarId:_ ToVarId:$0 GenChange:0}
		4:src.AssignVarOp{FromVarId:_ ToVarId:$1 GenChange:0}
		6:src.AssignVarOp{FromVarId:$2 ToVarId:$3 GenChange:1}
		8:src.AssignVarOp{FromVarId:$4 ToVarId:$0 GenChange:0}
		4:src.AssignVarOp{FromVarId:_ ToVarId:$1 GenChange:0}
		6:src.AssignVarOp{FromVarId:$2 ToVarId:$3 GenChange:1}
		8:src.AssignVarOp{FromVarId:$4 ToVarId:$0 GenChange:0}
	*/
	fset, funcDecl := utils.MustGenFunc(`func (s *Store[K, V]) processDeque(shard *Shard[K, V]) {
	send := make([]*Entry[K, V], 0, 2)
	for {
		shard.qlen -= int(evicted.cost.Load())
		send = append(send, evicted)
	}
}`)
	execution := ExecutionFromFunc(NewScopes(map[string]FuncId{
		SliceFuncName:  SliceFuncId,
		AppendFuncName: AppendFuncId,
	}), fset, funcDecl)
	t.Logf("%v", execution)
	/*
		0: 2:src.AssignVarOp{FromVarId:_ ToVarId:$0 GenChange:0}
		2: 1:src.NoOp{}
		1: 4:src.AssignVarOp{FromVarId:_ ToVarId:$1 GenChange:0}
		4: 3:src.NoOp{}
		3: 6:src.AssignVarOp{FromVarId:$2 ToVarId:$3 GenChange:1}
		6: 5:src.NoOp{}
		5: 8:src.AssignVarOp{FromVarId:$4 ToVarId:$0 GenChange:0}
		8: 7:src.NoOp{}
		7: 1:src.NoOp{}
	*/
	warnings := ValidateExecution(DefaultFuncSpecCollection, execution)
	for _, warning := range warnings {
		pos := execution.SourceCodeReferences.References[warning.ExecutionPoint]
		t.Logf("%v", execution.SourceCodeReferences.Fset.Position(pos))
	}
	require.Empty(t, warnings)
}

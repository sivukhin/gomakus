package src

import (
	"testing"

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
	simplified, _ := SimplifyExecution(SimplificationContext{Funcs: DefaultFuncSpecCollection}, execution)
	traces := GenerateTraces(simplified, 2)
	for i, trace := range traces {
		t.Logf("%v: %#v", i, ValidateTrace(trace))
	}
}

func TestTraceValidation(t *testing.T) {
	t.Log(ValidateTrace(ExecutionTrace{
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
	for _, warning := range warnings {
		t.Logf("warning found: %v", fset.Position(execution.SourceCodeReferences.References[warning.ExecutionPoint]))
	}
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
	for _, warning := range warnings {
		t.Logf("warning found: %v", fset.Position(execution.SourceCodeReferences.References[warning.ExecutionPoint]))
	}
}

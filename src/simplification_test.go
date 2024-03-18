package src

import (
	"testing"

	"gomakus/utils"
)

func TestSimplification(t *testing.T) {
	fset, funcDecl := utils.MustGenFunc(`func parseUnnestedKeyFieldSet(raw string, prefix []string) Set {
	ret := Set{}

	for _, s := range strings.Fields(raw) {
		next := append(prefix[:], s)
		ret = append(ret, next)
	}
	return ret
}`)
	execution := executionFromFunc(NewScopes(map[string]FuncId{
		SliceFuncName:  SliceFuncId,
		AppendFuncName: AppendFuncId,
	}), fset, funcDecl)
	t.Logf("%v", execution)
	simplified := SimplifyExecution(SimplificationContext{Funcs: DefaultFuncSpecCollection}, execution)
	t.Log(simplified)
}

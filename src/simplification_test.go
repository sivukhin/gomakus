package src

import (
	"testing"

	"github.com/sivukhin/gomakus/utils"
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
	execution := ExecutionFromFunc(NewScopes(map[string]FuncId{
		SliceFuncName:  SliceFuncId,
		AppendFuncName: AppendFuncId,
	}), fset, funcDecl)
	t.Logf("%v", execution)
	simplified, _ := SimplifyExecution(SimplificationContext{Funcs: DefaultFuncSpecCollection}, execution)
	t.Log(simplified)
}

func TestSimplification2(t *testing.T) {
	fset, funcDecl := utils.MustGenFunc(`func (r *FieldContext) Path() ast.Path {
	var path ast.Path
	for it := r; it != nil; it = it.Parent {
		if it.Index != nil {
			path = append(path, ast.PathIndex(*it.Index))
		} else if it.Field.Field != nil {
			path = append(path, ast.PathName(it.Field.Alias))
		}
	}

	// because we are walking up the chain, all the elements are backwards, do an inplace flip.
	for i := len(path)/2 - 1; i >= 0; i-- {
		opp := len(path) - 1 - i
		path[i], path[opp] = path[opp], path[i]
	}

	return path
}`)
	execution := ExecutionFromFunc(NewScopes(map[string]FuncId{
		SliceFuncName:  SliceFuncId,
		AppendFuncName: AppendFuncId,
	}), fset, funcDecl)
	t.Logf("%v", execution)
	simplified, _ := SimplifyExecution(SimplificationContext{Funcs: DefaultFuncSpecCollection}, execution)
	t.Log(simplified)
}

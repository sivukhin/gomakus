package src

import (
	"go/ast"
	"testing"

	"github.com/stretchr/testify/require"

	"gomakus/utils"
)

func TestIfDeconstruction(t *testing.T) {
	fset, statements := utils.MustGenStatements(`label: if x := 1; x == 1 {
panic("1")
} else if 2 == 2 {
panic("2")
} else if z := 3; 3 == 3 {
panic("3")
}else {
panic("4")
}`)
	ifStmt := utils.MustExtractLabeledStatement("label", statements...).(*ast.IfStmt)
	inits, bodies := deconstructIf(fset, ifStmt)
	require.Len(t, inits, 4)
	require.Len(t, bodies, 4)
	require.NotNil(t, inits[0])
	require.Nil(t, inits[1])
	require.NotNil(t, inits[2])
	require.Nil(t, inits[3])
}

func TestAssignment(t *testing.T) {
	{
		_, statements := utils.MustGenStatements(`label: a := 123`)
		assignStmt := utils.MustExtractLabeledStatement("label", statements...).(*ast.AssignStmt)
		t.Log(assignStmt.Tok)
	}
	{
		_, statements := utils.MustGenStatements(`label: a = 123`)
		assignStmt := utils.MustExtractLabeledStatement("label", statements...).(*ast.AssignStmt)
		t.Log(assignStmt.Tok)
	}
	{
		_, statements := utils.MustGenStatements(`label: a, b, c := f(1), f(2)`)
		assignStmt := utils.MustExtractLabeledStatement("label", statements...).(*ast.AssignStmt)
		t.Log(len(assignStmt.Lhs), len(assignStmt.Rhs))
	}
}

func TestDecl(t *testing.T) {
	{
		_, statements := utils.MustGenStatements(`label: var (
a, b int = 1, 2
c int = 3
)`)
		declStmt := utils.MustExtractLabeledStatement("label", statements...).(*ast.DeclStmt).Decl.(*ast.GenDecl)
		t.Log(declStmt.Tok)
		t.Log(len(declStmt.Specs))
	}
}

func TestSwitch(t *testing.T) {
	_, statements := utils.MustGenStatements(`label: switch 1 {
case 1: return
case 2: return
default: return
}`)
	switchStmt := utils.MustExtractLabeledStatement("label", statements...).(*ast.SwitchStmt)
	t.Logf("%#v", switchStmt.Tag)
}

func TestStructExpr(t *testing.T) {
	_, statements := utils.MustGenStatements(`
type t struct { x, y int }
type s struct { t }
a := t { x: 1, y: 2 }
label: x := s{ a }`)
	assignStmt := utils.MustExtractLabeledStatement("label", statements...).(*ast.AssignStmt)
	t.Logf("%T", assignStmt.Rhs[0].(*ast.CompositeLit).Elts[0])
}

func TestCallExpr(t *testing.T) {
	_, statements := utils.MustGenStatements(`
x := utils.Min([]string{}, "123")
label: x.print()`)
	assignStmt := utils.MustExtractLabeledStatement("label", statements...).(*ast.ExprStmt)
	t.Logf("%#v", assignStmt.X.(*ast.CallExpr))
}

func TestSelectorFunc(t *testing.T) {
	fset, funcDecl := utils.MustGenFunc(`func f(u *User) {
		u.Name = g()
}`)
	execution := ExecutionFromFunc(NewScopes(map[string]FuncId{
		SliceFuncName:  SliceFuncId,
		AppendFuncName: AppendFuncId,
		"g":            -3,
	}), fset, funcDecl)
	t.Logf("%v", execution)
}

func TestSelectors(t *testing.T) {
	fset, funcDecl := utils.MustGenFunc(`func f(a, b, c string) string {
	user := User {
		Name: a,
		Meta: Meta {
			Address: b,
			Phone: c,
		},
	}
	return user.Meta.Phone
}`)
	execution := ExecutionFromFunc(NewScopes(map[string]FuncId{
		SliceFuncName:  SliceFuncId,
		AppendFuncName: AppendFuncId,
	}), fset, funcDecl)
	t.Logf("%v", execution)
}

func TestMultiAssignment(t *testing.T) {
	fset, funcDecl := utils.MustGenFunc(`func f() string {
	a, b := f1()
	a, b = f2()
	return a + b
}`)
	execution := ExecutionFromFunc(NewScopes(map[string]FuncId{
		SliceFuncName:  SliceFuncId,
		AppendFuncName: AppendFuncId,
	}), fset, funcDecl)
	t.Logf("%v", execution)
}

func TestNakedReturn(t *testing.T) {
	fset, funcDecl := utils.MustGenFunc(`func f() (a int, b string) {
	a = 1
	b = "hi"
	return
}`)
	execution := ExecutionFromFunc(NewScopes(map[string]FuncId{
		SliceFuncName:  SliceFuncId,
		AppendFuncName: AppendFuncId,
	}), fset, funcDecl)
	t.Logf("%v", execution)
}

func TestBlocks(t *testing.T) {
	fset, funcDecl := utils.MustGenFunc(`func f() string {
	var result string
	{
		a := "hi"
		result = a
	}
	{
		a := "bye"
		result = a
	}
	return result
}`)
	execution := ExecutionFromFunc(NewScopes(map[string]FuncId{
		SliceFuncName:  SliceFuncId,
		AppendFuncName: AppendFuncId,
	}), fset, funcDecl)
	t.Logf("%v", execution)
}

func TestGqlgenBug(t *testing.T) {
	fset, funcDecl := utils.MustGenFunc(`func parseUnnestedKeyFieldSet(raw string, prefix []string) Set {
	ret := Set{}

	for _, s := range strings.Fields(raw) {
		next := append(prefix[:], s) //nolint:gocritic // slicing out on purpose
		ret = append(ret, next)
	}
	return ret
}`)
	execution := ExecutionFromFunc(NewScopes(map[string]FuncId{
		SliceFuncName:  SliceFuncId,
		AppendFuncName: AppendFuncId,
	}), fset, funcDecl)
	t.Logf("%v", execution)
}

func TestTheinegoBug(t *testing.T) {
	fset, funcDecl := utils.MustGenFunc(`func (s *Store[K, V]) processDeque(shard *Shard[K, V]) {
	if shard.qlen <= int(shard.qsize) {
		shard.mu.Unlock()
		return
	}
	// send to slru
	send := make([]*Entry[K, V], 0, 2)
	// removed because frequency < slru tail frequency
	removedkv := make([]dequeKV[K, V], 0, 2)
	// expired
	expiredkv := make([]dequeKV[K, V], 0, 2)
	// expired
	for shard.qlen > int(shard.qsize) {
		evicted := shard.deque.PopBack()
		evicted.deque = false
		expire := evicted.expire.Load()
		shard.qlen -= int(evicted.cost.Load())
		if expire != 0 && expire <= s.timerwheel.clock.NowNano() {
			deleted := shard.delete(evicted)
			// double check because entry maybe removed already by Delete API
			if deleted {
				expiredkv = append(expiredkv, s.kvBuilder(evicted))
				s.postDelete(evicted)
			}
		} else {
			count := evicted.frequency.Load()
			threshold := s.policy.threshold.Load()
			if count == -1 {
				send = append(send, evicted)
			} else {
				if int32(count) >= threshold {
					send = append(send, evicted)
				} else {
					deleted := shard.delete(evicted)
					// double check because entry maybe removed already by Delete API
					if deleted {
						removedkv = append(
							expiredkv, s.kvBuilder(evicted),
						)
						s.postDelete(evicted)
					}
				}
			}
		}
	}
	shard.mu.Unlock()
	for _, entry := range send {
		s.writebuf <- WriteBufItem[K, V]{entry: entry, code: NEW}
	}
	if s.removalListener != nil {
		for _, e := range removedkv {
			_ = s.removalCallback(e, EVICTED)
		}
		for _, e := range expiredkv {
			_ = s.removalCallback(e, EXPIRED)
		}
	}
}`)
	execution := ExecutionFromFunc(NewScopes(map[string]FuncId{
		SliceFuncName:  SliceFuncId,
		AppendFuncName: AppendFuncId,
	}), fset, funcDecl)
	t.Logf("%v", execution)
}

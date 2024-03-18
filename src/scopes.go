package src

import (
	"fmt"
)

const (
	SliceFuncId    FuncId = -1
	SliceFuncName  string = "$Slice"
	AppendFuncId   FuncId = -2
	AppendFuncName string = "append"
)

type Scopes struct {
	Funcs     map[string]FuncId
	Vars      []map[string]VarId
	LastVarId *VarId
}

func NewScopes(funcs map[string]FuncId) Scopes {
	vars := []map[string]VarId{make(map[string]VarId)}
	lastVarId := VarId(0)
	return Scopes{
		Funcs:     funcs,
		Vars:      vars,
		LastVarId: &lastVarId,
	}
}

func (s Scopes) NewVarId() VarId {
	newId := *s.LastVarId
	*s.LastVarId++
	return newId
}

func (s Scopes) TryGetFunc(name string) (FuncId, bool) {
	f, ok := s.Funcs[name]
	return f, ok
}

func (s Scopes) GetFunc(name string) FuncId {
	f, ok := s.Funcs[name]
	if !ok {
		panic(fmt.Errorf("unable to find func '%v' in the scope", name))
	}
	return f
}

func (s Scopes) GetVarOrBlank(name string) VarId {
	for i := len(s.Vars) - 1; i >= 0; i-- {
		if id, ok := s.Vars[i][name]; ok {
			return id
		}
	}
	return BlankVarId
}

func (s Scopes) GetVar(name string) VarId {
	if name == BlankVarName {
		return BlankVarId
	}
	for i := len(s.Vars) - 1; i >= 0; i-- {
		if id, ok := s.Vars[i][name]; ok {
			return id
		}
	}
	panic(fmt.Errorf("variable '%v' not found", name))
}

func (s Scopes) CreateVar(name string) VarId {
	if name == BlankVarName {
		return BlankVarId
	}
	newId := s.NewVarId()
	s.Vars[len(s.Vars)-1][name] = newId
	return newId
}

func (s Scopes) PushScope() Scopes {
	return Scopes{
		Funcs:     s.Funcs,
		Vars:      append(s.Vars, make(map[string]VarId)),
		LastVarId: s.LastVarId,
	}
}

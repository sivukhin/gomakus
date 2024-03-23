package src

import (
	"fmt"
	"go/token"
	"strconv"
	"strings"
)

type (
	VarId int
	// ExecutionPoint uniquely identifies code point within single function
	ExecutionPoint int
	// ExecutionTransition represents single operation which change IR state
	ExecutionTransition struct {
		ToPoint   ExecutionPoint
		Operation Operation
	}
	// Execution represents control flow graph with all possible transitions of the single function
	Execution struct {
		RootPoint            ExecutionPoint
		Transitions          map[ExecutionPoint][]ExecutionTransition
		SourceCodeReferences SourceCodeReferences
	}
	SourceCodeReferences struct {
		Fset       *token.FileSet
		References map[ExecutionPoint]token.Pos
	}
)

func (v VarId) String() string {
	if v == BlankVarId {
		return BlankVarName
	}
	return "$" + strconv.Itoa(int(v))
}
func (v VarSelector) String() string {
	return strings.Join(append([]string{v.VarId.String()}, v.Selector...), ".")
}
func (v VarEmbed) String() string {
	return strings.Join(append(append([]string{}, v.Path...), v.VarSelector.String()), ":")
}

func (e ExecutionTransition) String() string {
	return fmt.Sprintf("%v:%T%+v ", e.ToPoint, e.Operation, e.Operation)
}

func (e Execution) String() string {
	var s strings.Builder
	s.WriteString(fmt.Sprintf("execution (%v):\n", e.RootPoint))
	edges := make(map[ExecutionPoint][]ExecutionPoint)
	for point, transitions := range e.Transitions {
		for _, transition := range transitions {
			edges[point] = append(edges[point], transition.ToPoint)
		}
	}
	points := TopologyOrder(e.RootPoint, func(v ExecutionPoint) []ExecutionPoint { return edges[v] })
	for _, point := range points {
		if pos, ok := e.SourceCodeReferences.References[point]; e.SourceCodeReferences.Fset != nil && ok {
			ref := e.SourceCodeReferences.Fset.Position(pos)
			s.WriteString(fmt.Sprintf("%v(%v[%v])::\t", ref.Filename, ref.Line, ref.Column))
		}
		s.WriteString(fmt.Sprintf("  %v: ", point))
		for _, next := range e.Transitions[point] {
			s.WriteString(fmt.Sprintf("%v:%T%+v ", next.ToPoint, next.Operation, next.Operation))
		}
		s.WriteString("\n")
	}
	return s.String()
}

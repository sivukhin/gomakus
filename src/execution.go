package src

import (
	"fmt"
	"go/token"
	"sort"
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
		Enabled    bool
		References map[ExecutionPoint]token.Position
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

func (e Execution) String() string {
	var s strings.Builder
	s.WriteString(fmt.Sprintf("execution (%v):\n", e.RootPoint))
	points := make([]ExecutionPoint, 0)
	for point := range e.Transitions {
		points = append(points, point)
	}
	sort.Slice(points, func(i, j int) bool {
		refI := e.SourceCodeReferences.References[points[i]]
		refJ := e.SourceCodeReferences.References[points[j]]
		if refI.Filename != refJ.Filename {
			return refI.Filename < refJ.Filename
		} else if refI.Line != refJ.Line {
			return refI.Line < refJ.Line
		} else if refI.Column != refJ.Column {
			return refI.Column < refJ.Column
		} else {
			return points[i] < points[j]
		}
	})
	for _, point := range points {
		if ref, ok := e.SourceCodeReferences.References[point]; ok {
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

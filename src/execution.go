package src

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type (
	// ExecutionPoint uniquely identifies code point within single function
	ExecutionPoint int
	// ExecutionTransition represents single operation which change IR state
	ExecutionTransition struct {
		ToPoint   ExecutionPoint
		Operation Operation
	}
	// Execution represents control flow graph with all possible transitions of the single function
	Execution struct {
		RootPoint   ExecutionPoint
		Transitions map[ExecutionPoint][]ExecutionTransition
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
	transitions := make([]ExecutionPoint, 0)
	for point := range e.Transitions {
		transitions = append(transitions, point)
	}
	sort.Slice(transitions, func(i, j int) bool { return transitions[i] < transitions[j] })
	for _, point := range transitions {
		s.WriteString(fmt.Sprintf("  %v: ", point))
		for _, next := range e.Transitions[point] {
			s.WriteString(fmt.Sprintf("%v:%T%+v ", next.ToPoint, next.Operation, next.Operation))
		}
		s.WriteString("\n")
	}
	return s.String()
}

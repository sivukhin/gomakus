package src

import (
	"fmt"
	"slices"
	"strings"
)

type ExecutionTrace []ExecutionTransition

func (t ExecutionTrace) String() string {
	var s strings.Builder
	for _, e := range t {
		s.WriteString(fmt.Sprintf("%v\n", e))
	}
	return s.String()
}

func GenerateTraces(execution Execution, repeatLimit int) []ExecutionTrace {
	visits := make(map[ExecutionPoint]int)
	traces := make([]ExecutionTrace, 0)
	generateTraces(execution, execution.RootPoint, repeatLimit, nil, visits, &traces)
	return traces
}

func generateTraces(
	execution Execution,
	root ExecutionPoint,
	repeatLimit int,
	current ExecutionTrace,
	visits map[ExecutionPoint]int,
	traces *[]ExecutionTrace,
) {
	finish := true
	for _, transition := range execution.Transitions[root] {
		if visits[transition.ToPoint] >= repeatLimit {
			continue
		}
		visits[transition.ToPoint]++
		generateTraces(execution, transition.ToPoint, repeatLimit, append(current, transition), visits, traces)
		visits[transition.ToPoint]--
		finish = false
	}
	if finish {
		*traces = append(*traces, slices.Clone(current))
	}
}

type ValidationWarning struct {
	Trace          ExecutionTrace
	ExecutionPoint ExecutionPoint
}

func ValidateExecution(funcs map[FuncId]FuncSpec, execution Execution) []ValidationWarning {
	simplified, simplifiedToOriginal := SimplifyExecution(SimplificationContext{Funcs: funcs}, execution)
	traces := GenerateTraces(simplified, 2)
	warnedExecutionPoints := make(map[ExecutionPoint]struct{})
	var warnings []ValidationWarning
	for _, trace := range traces {
		for _, warning := range ValidateTrace(trace) {
			if _, ok := warnedExecutionPoints[warning.ExecutionPoint]; ok {
				continue
			}
			warnedExecutionPoints[warning.ExecutionPoint] = struct{}{}
			warnings = append(warnings, ValidationWarning{
				Trace:          trace,
				ExecutionPoint: simplifiedToOriginal[warning.ExecutionPoint],
			})
		}
	}
	return warnings
}

func ValidateTrace(trace ExecutionTrace) []ValidationWarning {
	originLatestGen := make(map[int]int, 0)
	variableGen := make(map[VarId]VarGen)
	warnings := make([]ValidationWarning, 0)
	valueId := 0
	for _, transition := range trace {
		switch statement := transition.Operation.(type) {
		case AssignVarOp:
			if statement.ToVarId == BlankVarId {
				continue
			}
			var targetGen VarGen
			if statement.FromVarId == BlankVarId {
				targetGen = VarGen{Id: valueId}
				valueId++
			} else {
				sourceGen, ok := variableGen[statement.FromVarId]
				if !ok {
					sourceGen = VarGen{Id: valueId}
					variableGen[statement.FromVarId] = VarGen{Id: valueId}
					valueId++
				}
				targetGen = VarGen{
					Id:  sourceGen.Id,
					Gen: sourceGen.Gen + int(statement.GenChange),
				}
			}
			latestGen, ok := originLatestGen[targetGen.Id]
			if ok && statement.GenChange == NextGen && targetGen.Gen <= latestGen {
				err := ValidationWarning{
					Trace:          trace,
					ExecutionPoint: transition.ToPoint,
				}
				warnings = append(warnings, err)
			} else if !ok || targetGen.Gen > latestGen {
				originLatestGen[targetGen.Id] = targetGen.Gen
			}
			variableGen[statement.ToVarId] = targetGen
		case NoOp:
			continue
		default:
			panic(fmt.Errorf("unexpected execution statement type(%T): %#v", transition, transition))
		}
	}
	return warnings
}

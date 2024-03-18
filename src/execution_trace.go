package src

import (
	"fmt"
	"slices"
)

type ExecutionTrace []ExecutionTransition

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
	Trace            ExecutionTrace
	ExecutionPoint   ExecutionPoint
	ConflictOrigin   VarId
	ConflictVariable VarId
}

func ValidateTrace(trace ExecutionTrace) []ValidationWarning {
	originLatestGen := make(map[VarId]int, 0)
	variableGen := make(map[VarId]VarGen)
	errs := make([]ValidationWarning, 0)
	for _, transition := range trace {
		switch statement := transition.Operation.(type) {
		case AssignVarOp:
			sourceGen, ok := variableGen[statement.FromVarId]
			var targetGen VarGen
			if !ok {
				targetGen = VarGen{
					Id:  statement.FromVarId,
					Gen: 0,
				}
			} else {
				targetGen = VarGen{
					Id:  sourceGen.Id,
					Gen: sourceGen.Gen + int(statement.GenChange),
				}
			}
			latestGen, ok := originLatestGen[targetGen.Id]
			if ok && statement.GenChange == NextGen && targetGen.Gen <= latestGen {
				err := ValidationWarning{
					Trace:            trace,
					ExecutionPoint:   transition.ToPoint,
					ConflictOrigin:   targetGen.Id,
					ConflictVariable: statement.FromVarId,
				}
				errs = append(errs, err)
			} else if !ok || latestGen < targetGen.Gen {
				originLatestGen[targetGen.Id] = targetGen.Gen
			}
			variableGen[statement.ToVarId] = targetGen
		case NoOp:
			continue
		default:
			panic(fmt.Errorf("unexpected execution statement type(%T): %#v", transition, transition))
		}
	}
	return errs
}

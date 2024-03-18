package src

import (
	"fmt"
	"slices"
)

type (
	VarId         int
	StatementId   int
	GenChangeType int
	VariableGen   struct {
		Id  VarId
		Gen int
	}
)

const (
	PrevGen GenChangeType = -1
	SameGen GenChangeType = 0
	NextGen GenChangeType = +1
)

type (
	NopStatement        struct{}
	UseStatement        struct{ Values []VarId }
	AssignmentStatement struct {
		Target, Source VarId
		GenChange      GenChangeType
	}

	ExecutionStatement struct {
		Id        StatementId
		Statement any
	}
)

type ExecutionGraphCursor struct {
	Graph   *ExecutionGraph
	Current StatementId
}

func NewExecutionGraphCursor(graph *ExecutionGraph) ExecutionGraphCursor {
	current := graph.NewNode(NopStatement{})
	return ExecutionGraphCursor{Graph: graph, Current: current}
}
func (c *ExecutionGraphCursor) Next(statement any) {
	next := c.Graph.NewNode(statement)
	c.Graph.AddEdge(c.Current, next)
	c.Current = next
}

type ExecutionGraph struct {
	Nodes map[StatementId]any
	Edges map[StatementId][]StatementId
}

func NewExecutionGraph() *ExecutionGraph {
	return &ExecutionGraph{
		Nodes: make(map[StatementId]any),
		Edges: make(map[StatementId][]StatementId),
	}
}

func (graph *ExecutionGraph) NewNode(statement any) StatementId {
	id := StatementId(len(graph.Nodes))
	graph.Nodes[id] = statement
	return id
}
func (graph *ExecutionGraph) AddEdge(source, target StatementId) {
	graph.Edges[source] = append(graph.Edges[source], target)
}

type ExecutionError struct {
	Trace            ExecutionTrace
	Id               StatementId
	ConflictOrigin   VarId
	ConflictVariable VarId
}

type ExecutionTrace []ExecutionStatement
type ExecutionContext struct{ NextGenArguments map[StatementId][]int }

// todo (sivukhin, 2023-11-28): implement generator
func (graph *ExecutionGraph) GenerateTraces(root StatementId, repeatLimit int) []ExecutionTrace {
	visits := make(map[StatementId]int)
	traces := make([]ExecutionTrace, 0)
	graph.generateTraces(root, repeatLimit, nil, visits, &traces)
	return traces
}

func (graph *ExecutionGraph) generateTraces(
	root StatementId,
	repeatLimit int,
	current ExecutionTrace,
	visits map[StatementId]int,
	traces *[]ExecutionTrace,
) {
	statement := ExecutionStatement{Id: root, Statement: graph.Nodes[root]}
	current = append(current, statement)
	finish := true
	for _, next := range graph.Edges[root] {
		if visits[next] >= repeatLimit {
			continue
		}
		visits[next]++
		graph.generateTraces(next, repeatLimit, current, visits, traces)
		visits[next]--
		finish = false
	}
	if finish {
		*traces = append(*traces, slices.Clone(current))
	} else if true {

	} else {

	}
}

func (ctx ExecutionContext) ValidateGraph(graph *ExecutionGraph, root StatementId, repeatLimit int) []ExecutionError {
	errs := make([]ExecutionError, 0)
	for _, trace := range graph.GenerateTraces(root, repeatLimit) {
		errs = append(errs, ctx.ValidateTrace(trace)...)
	}
	return errs
}

func (ctx ExecutionContext) ValidateTrace(trace ExecutionTrace) []ExecutionError {
	originLatestGen := make(map[VarId]int, 0)
	variableGen := make(map[VarId]VariableGen)
	errs := make([]ExecutionError, 0)
	for _, someStatement := range trace {
		switch statement := someStatement.Statement.(type) {
		case AssignmentStatement:
			sourceGen, ok := variableGen[statement.Source]
			var targetGen VariableGen
			if !ok {
				targetGen = VariableGen{
					Id:  statement.Source,
					Gen: 0,
				}
			} else {
				targetGen = VariableGen{
					Id:  sourceGen.Id,
					Gen: sourceGen.Gen + int(statement.GenChange),
				}
			}
			latestGen, ok := originLatestGen[targetGen.Id]
			if ok && statement.GenChange == NextGen && targetGen.Gen <= latestGen {
				err := ExecutionError{
					Trace:            trace,
					Id:               someStatement.Id,
					ConflictOrigin:   targetGen.Id,
					ConflictVariable: statement.Source,
				}
				errs = append(errs, err)
			} else if !ok || latestGen < targetGen.Gen {
				originLatestGen[targetGen.Id] = targetGen.Gen
			}
			variableGen[statement.Target] = targetGen
		case UseStatement:
			nextGenArguments := ctx.NextGenArguments[someStatement.Id]
			for _, position := range nextGenArguments {
				origin := variableGen[statement.Values[position]]
				latestGen, ok := originLatestGen[origin.Id]
				if ok && origin.Gen < latestGen {
					err := ExecutionError{
						Trace:            trace,
						Id:               someStatement.Id,
						ConflictOrigin:   origin.Id,
						ConflictVariable: statement.Values[position],
					}
					errs = append(errs, err)
				}
			}
		case NopStatement:
			continue
		default:
			panic(fmt.Errorf("unexpected execution statement type(%T): %#v", someStatement, someStatement))
		}
	}
	return errs
}

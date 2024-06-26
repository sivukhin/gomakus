package src

import (
	"go/token"
)

type ExecutionBuilder struct {
	CurrentPoint ExecutionPoint
	lastPoint    *ExecutionPoint
	execution    *Execution
}

func NewExecutionBuilder(fset *token.FileSet) ExecutionBuilder {
	rootPoint := ExecutionPoint(0)
	lastPoint := ExecutionPoint(0)
	transitions := make(map[ExecutionPoint][]ExecutionTransition)
	return ExecutionBuilder{
		CurrentPoint: rootPoint,
		lastPoint:    &lastPoint,
		execution: &Execution{
			RootPoint:   rootPoint,
			Transitions: transitions,
			SourceCodeReferences: SourceCodeReferences{
				Fset:       fset,
				References: make(map[ExecutionPoint]token.Pos),
			},
		},
	}
}

func (b ExecutionBuilder) Build() Execution { return *b.execution }

func (b ExecutionBuilder) AssignRef(point ExecutionPoint, position token.Pos) {
	if (*b.execution).SourceCodeReferences.Fset != nil {
		(*b.execution).SourceCodeReferences.References[point] = position
	}
}

func (b ExecutionBuilder) AcquirePoint() ExecutionBuilder {
	*b.lastPoint += 1
	return ExecutionBuilder{
		CurrentPoint: *b.lastPoint,
		lastPoint:    b.lastPoint,
		execution:    b.execution,
	}
}

func (b ExecutionBuilder) AcquirePointWithRef(position token.Pos) ExecutionBuilder {
	*b.lastPoint += 1
	b.AssignRef(*b.lastPoint, position)
	return ExecutionBuilder{
		CurrentPoint: *b.lastPoint,
		lastPoint:    b.lastPoint,
		execution:    b.execution,
	}
}

func (b ExecutionBuilder) ConnectTo(point ExecutionPoint) ExecutionBuilder {
	b.connect(ExecutionTransition{ToPoint: point, Operation: NoOp{}})
	return ExecutionBuilder{
		CurrentPoint: point,
		lastPoint:    b.lastPoint,
		execution:    b.execution,
	}
}

func (b ExecutionBuilder) ApplyNext(op Operation) ExecutionBuilder {
	next := b.AcquirePoint()
	b.connect(ExecutionTransition{ToPoint: next.CurrentPoint, Operation: op})
	return ExecutionBuilder{
		CurrentPoint: next.CurrentPoint,
		lastPoint:    b.lastPoint,
		execution:    b.execution,
	}
}

func (b ExecutionBuilder) ApplyNextWithRef(op Operation, position token.Pos) ExecutionBuilder {
	next := b.AcquirePointWithRef(position)
	b.connect(ExecutionTransition{ToPoint: next.CurrentPoint, Operation: op})
	return ExecutionBuilder{
		CurrentPoint: next.CurrentPoint,
		lastPoint:    b.lastPoint,
		execution:    b.execution,
	}
}

func (b ExecutionBuilder) connect(transition ExecutionTransition) {
	(*b.execution).Transitions[b.CurrentPoint] = append((*b.execution).Transitions[b.CurrentPoint], transition)
}

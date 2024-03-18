package src

type ExecutionBuilder struct {
	CurrentPoint ExecutionPoint
	lastPoint    *ExecutionPoint
	execution    *Execution
}

func NewExecutionBuilder() ExecutionBuilder {
	rootPoint := ExecutionPoint(0)
	lastPoint := ExecutionPoint(0)
	transitions := make(map[ExecutionPoint][]ExecutionTransition)
	return ExecutionBuilder{
		CurrentPoint: rootPoint,
		lastPoint:    &lastPoint,
		execution:    &Execution{RootPoint: rootPoint, Transitions: transitions},
	}
}

func (b ExecutionBuilder) Build() Execution { return *b.execution }

func (b ExecutionBuilder) AcquirePoint() ExecutionBuilder {
	*b.lastPoint += 1
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

func (b ExecutionBuilder) connect(transition ExecutionTransition) {
	(*b.execution).Transitions[b.CurrentPoint] = append((*b.execution).Transitions[b.CurrentPoint], transition)
}

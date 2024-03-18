package src

import (
	"testing"
)

func TestBuilderApi(t *testing.T) {
	builder := NewExecutionBuilder()
	builder = builder.ApplyNext(AssignVarOp{ToVarId: 0, FromVarId: BlankVarId})
	builder = builder.ApplyNext(AssignVarOp{ToVarId: 1, FromVarId: BlankVarId})
	afterStateId := builder.AcquirePoint()
	builder.ApplyNext(AssignVarOp{ToVarId: 2, FromVarId: BlankVarId}).ConnectTo(afterStateId.CurrentPoint)
	builder.ApplyNext(AssignVarOp{ToVarId: 3, FromVarId: BlankVarId}).ConnectTo(afterStateId.CurrentPoint)
	t.Logf("%+v", builder.Build())
}

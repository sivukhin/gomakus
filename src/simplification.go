package src

import (
	"go-append-check/utils"
)

type SimplificationContext struct {
	Funcs map[FuncId]FuncSpec
}

func SimplifyExecution(context SimplificationContext, execution Execution) Execution {
	assigns := make([]AssignSelectorOp, 0)
	for _, transitions := range execution.Transitions {
		for _, transition := range transitions {
			switch op := transition.Operation.(type) {
			case AssignSelectorOp:
				assigns = append(assigns, op)
			case UseSelectorsOp:
				funcSpec, ok := context.Funcs[op.FuncId]
				if !ok {
					continue
				}
				for _, output := range funcSpec.Outputs {
					for i, composition := range output {
						for _, returnRef := range composition {
							assigns = append(assigns, AssignSelectorOp{
								FromSelector: returnRef.InputArgEmbed.VarSelector,
								ToSelector:   VarSelector{VarId: op.Outputs[i], Selector: returnRef.InputArgEmbed.Path},
							})
						}
					}
				}
			}
		}
	}

	factorization := FactorizeAssignments(assigns)
	builder := NewExecutionBuilder()
	_ = factorization
	_ = builder
	//varSelectorCollection := NewVarSelectorCollection()
	//simplifyExecution(factorization, builder, execution.RootPoint, execution)
	panic("!")
}

type simplificationContext struct {
	funcs                    map[FuncId]FuncSpec
	factorization            FactorizationRules
	execution                Execution
	varSelectorCollection    varSelectorCollection
	executionPointCollection executionPointCollection
	visited                  map[ExecutionPoint]struct{}
}

func (c simplificationContext) simplifyExecution(builder ExecutionBuilder, point ExecutionPoint) {
	_, ok := c.visited[point]
	if ok {
		return
	}
	c.visited[point] = struct{}{}

	transitions, ok := c.execution.Transitions[point]
	if !ok {
		return
	}
	for _, transition := range transitions {
		toPoint := c.executionPointCollection.AcquireOrGet(builder, transition.ToPoint)
		switch operation := transition.Operation.(type) {
		case AssignSelectorOp:
			fromSelectors := c.factorization.FactorizeSelector(operation.FromSelector)
			toSelectors := c.factorization.FactorizeSelector(operation.ToSelector)
			utils.Assertf(len(fromSelectors) == len(toSelectors), "inconsistent assignment operator factorization: from=%+v, to=%+v", fromSelectors, toSelectors)
			for i := range fromSelectors {
				fromVar := c.varSelectorCollection.IntroduceVarOrGet(fromSelectors[i])
				toVar := c.varSelectorCollection.IntroduceVarOrGet(toSelectors[i])
				builder = builder.ApplyNext(AssignVarOp{FromVarId: fromVar, ToVarId: toVar})
			}
		case UseSelectorsOp:

		}
		builder.ConnectTo(toPoint)
	}
}

type varSelectorCollection map[string]int

func (c varSelectorCollection) IntroduceVarOrGet(selector VarSelector) VarId {
	selectorString := selector.String()
	varId, ok := c[selectorString]
	if ok {
		return VarId(varId)
	}
	lastVarId := len(c)
	c[selectorString] = lastVarId
	return VarId(lastVarId)
}

type executionPointCollection map[ExecutionPoint]ExecutionPoint

func (c executionPointCollection) AcquireOrGet(builder ExecutionBuilder, point ExecutionPoint) ExecutionPoint {
	if _, ok := c[point]; !ok {
		c[point] = builder.AcquirePoint().CurrentPoint
	}
	return c[point]
}

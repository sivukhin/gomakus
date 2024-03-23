package src

import (
	"gomakus/utils"
)

type SimplificationContext struct {
	Funcs map[FuncId]FuncSpec
}

func SelectAssignOps(context SimplificationContext, execution Execution) []AssignSelectorOp {
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
					for i, outputComponent := range output {
						assigns = append(assigns, AssignSelectorOp{
							// todo (sivukhin, 2024-03-23): very dirty hack!
							FromSelector: op.Inputs[outputComponent.InputRef.ArgIndex][outputComponent.InputRef.SelectorIndex].VarSelector,
							ToSelector:   VarSelector{VarId: op.Outputs[i], Selector: outputComponent.OutputPath},
						})
					}
				}
			}
		}
	}
	return assigns
}

func SimplifyExecution(context SimplificationContext, execution Execution) (Execution, map[ExecutionPoint]ExecutionPoint) {
	assigns := SelectAssignOps(context, execution)
	simplification := &simplificationContext{
		funcs:                    context.Funcs,
		factorization:            FactorizeAssignments(assigns),
		varSelectorCollection:    make(varSelectorCollection),
		executionPointCollection: make(executionPointCollection),
		visited:                  make(map[ExecutionPoint]struct{}),
		simplifiedToOriginal:     make(map[ExecutionPoint]ExecutionPoint),
	}
	builder := NewExecutionBuilder(nil)
	simplification.simplifyExecution(builder, execution, execution.RootPoint)
	return builder.Build(), simplification.simplifiedToOriginal
}

type simplificationContext struct {
	funcs                    map[FuncId]FuncSpec
	factorization            FactorizationRules
	varSelectorCollection    varSelectorCollection
	executionPointCollection executionPointCollection
	visited                  map[ExecutionPoint]struct{}
	simplifiedToOriginal     map[ExecutionPoint]ExecutionPoint
}

func (c *simplificationContext) simplifyExecution(builder ExecutionBuilder, execution Execution, point ExecutionPoint) {
	_, ok := c.visited[point]
	if ok {
		return
	}
	c.visited[point] = struct{}{}
	c.simplifiedToOriginal[builder.CurrentPoint] = point

	transitions, ok := execution.Transitions[point]
	if !ok {
		return
	}
	original := builder
	for _, transition := range transitions {
		builder = original
		toPoint := c.executionPointCollection.AcquireOrGet(builder, transition.ToPoint)
		switch operation := transition.Operation.(type) {
		case AssignSelectorOp:
			if operation.FromSelector.VarId == BlankVarId && operation.ToSelector.VarId != BlankVarId {
				toSelectors := c.factorization.FactorizeSelector(operation.ToSelector)
				for i := range toSelectors {
					fromVar := c.varSelectorCollection.IntroduceVarOrGet(VarSelector{VarId: BlankVarId})
					toVar := c.varSelectorCollection.IntroduceVarOrGet(toSelectors[i])

					builder = builder.ApplyNext(AssignVarOp{FromVarId: fromVar, ToVarId: toVar})
					c.simplifiedToOriginal[builder.CurrentPoint] = transition.ToPoint
				}
			} else if operation.FromSelector.VarId != BlankVarId && operation.ToSelector.VarId == BlankVarId {
				fromSelectors := c.factorization.FactorizeSelector(operation.FromSelector)
				for i := range fromSelectors {
					fromVar := c.varSelectorCollection.IntroduceVarOrGet(fromSelectors[i])
					toVar := c.varSelectorCollection.IntroduceVarOrGet(VarSelector{VarId: BlankVarId})

					builder = builder.ApplyNext(AssignVarOp{FromVarId: fromVar, ToVarId: toVar})
					c.simplifiedToOriginal[builder.CurrentPoint] = transition.ToPoint
				}
			} else {
				fromSelectors := c.factorization.FactorizeSelector(operation.FromSelector)
				toSelectors := c.factorization.FactorizeSelector(operation.ToSelector)
				utils.Assertf(len(fromSelectors) == len(toSelectors), "inconsistent assignment operator factorization: from=%+v, to=%+v", fromSelectors, toSelectors)
				for i := range fromSelectors {
					fromVar := c.varSelectorCollection.IntroduceVarOrGet(fromSelectors[i])
					toVar := c.varSelectorCollection.IntroduceVarOrGet(toSelectors[i])

					builder = builder.ApplyNext(AssignVarOp{FromVarId: fromVar, ToVarId: toVar})
					c.simplifiedToOriginal[builder.CurrentPoint] = transition.ToPoint
				}
			}
		case UseSelectorsOp:
			if funcSpec, ok := c.funcs[operation.FuncId]; ok {
				utils.Assertf(len(funcSpec.Outputs) == len(operation.Outputs), "spec outputs must have same length as operation outputs: %v != %v", len(funcSpec.Outputs), len(operation.Outputs))
				for i, output := range operation.Outputs {
					funcOutputComponents := funcSpec.Outputs[i]
					for _, funcInputRef := range funcOutputComponents {
						var fromSelector, toSelector VarSelector
						if funcInputRef.InputRef.ArgIndex == BlankVarId {
							fromSelector = VarSelector{VarId: BlankVarId}
							toSelector = VarSelector{VarId: output, Selector: funcInputRef.OutputPath}
						} else {
							inputSelector := funcSpec.Inputs[funcInputRef.InputRef.ArgIndex][funcInputRef.InputRef.SelectorIndex].Selector
							utils.Assertf(len(inputSelector) == 0, "only degenerate selectors are supported: %#v", inputSelector)

							for _, inputEmbed := range operation.Inputs[funcInputRef.InputRef.ArgIndex] {
								fromSelector = VarSelector{VarId: inputEmbed.VarSelector.VarId, Selector: inputEmbed.VarSelector.Selector}
								toSelector = VarSelector{VarId: output, Selector: append(append(Path{}, funcInputRef.OutputPath...), inputEmbed.Path...)}
							}
						}
						fromVar := c.varSelectorCollection.IntroduceVarOrGet(fromSelector)
						toVar := c.varSelectorCollection.IntroduceVarOrGet(toSelector)
						builder = builder.ApplyNext(AssignVarOp{FromVarId: fromVar, ToVarId: toVar, GenChange: funcInputRef.GenChange})
						c.simplifiedToOriginal[builder.CurrentPoint] = transition.ToPoint
					}
				}
			} else {
				for _, output := range operation.Outputs {
					builder = builder.ApplyNext(AssignVarOp{FromVarId: output, ToVarId: BlankVarId})
					c.simplifiedToOriginal[builder.CurrentPoint] = transition.ToPoint
				}
			}
		}
		c.simplifyExecution(builder.ConnectTo(toPoint), execution, transition.ToPoint)
	}
}

type varSelectorCollection map[string]int

func (c varSelectorCollection) IntroduceVarOrGet(selector VarSelector) VarId {
	if selector.VarId == BlankVarId {
		return BlankVarId
	}
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

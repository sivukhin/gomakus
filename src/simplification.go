package src

import (
	"gomakus/utils"
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
					for i, outputComponent := range output {
						assigns = append(assigns, AssignSelectorOp{
							FromSelector: funcSpec.Inputs[outputComponent.InputRef.ArgIndex][outputComponent.InputRef.SelectorIndex],
							ToSelector:   VarSelector{VarId: op.Outputs[i], Selector: outputComponent.OutputPath},
						})
					}
				}
			}
		}
	}

	simplification := &simplificationContext{
		funcs:                    context.Funcs,
		factorization:            FactorizeAssignments(assigns),
		varSelectorCollection:    make(varSelectorCollection),
		executionPointCollection: make(executionPointCollection),
		visited:                  make(map[ExecutionPoint]struct{}),
	}
	builder := NewExecutionBuilder(execution.SourceCodeReferences.Enabled)
	builder.AssignRef(builder.CurrentPoint, execution.SourceCodeReferences.References[execution.RootPoint])

	simplification.simplifyExecution(builder, execution, execution.RootPoint)
	return builder.Build()
}

type simplificationContext struct {
	funcs                    map[FuncId]FuncSpec
	factorization            FactorizationRules
	varSelectorCollection    varSelectorCollection
	executionPointCollection executionPointCollection
	visited                  map[ExecutionPoint]struct{}
}

func (c *simplificationContext) simplifyExecution(builder ExecutionBuilder, execution Execution, point ExecutionPoint) {
	_, ok := c.visited[point]
	if ok {
		return
	}
	c.visited[point] = struct{}{}

	ref := execution.SourceCodeReferences.References[point]
	builder.AssignRef(builder.CurrentPoint, ref)

	transitions, ok := execution.Transitions[point]
	if !ok {
		return
	}
	for _, transition := range transitions {
		toPoint := c.executionPointCollection.AcquireOrGet(builder, transition.ToPoint)
		switch operation := transition.Operation.(type) {
		case AssignVarOp:
		case AssignSelectorOp:
			fromSelectors := c.factorization.FactorizeSelector(operation.FromSelector)
			toSelectors := c.factorization.FactorizeSelector(operation.ToSelector)
			utils.Assertf(len(fromSelectors) == len(toSelectors), "inconsistent assignment operator factorization: from=%+v, to=%+v", fromSelectors, toSelectors)
			for i := range fromSelectors {
				fromVar := c.varSelectorCollection.IntroduceVarOrGet(fromSelectors[i])
				toVar := c.varSelectorCollection.IntroduceVarOrGet(toSelectors[i])
				builder = builder.ApplyNextWithRef(AssignVarOp{FromVarId: fromVar, ToVarId: toVar}, ref)
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
						builder = builder.ApplyNextWithRef(AssignVarOp{FromVarId: fromVar, ToVarId: toVar, GenChange: funcInputRef.GenChange}, ref)
					}
				}
			} else {
				for _, output := range operation.Outputs {
					builder = builder.ApplyNextWithRef(AssignVarOp{FromVarId: output, ToVarId: BlankVarId}, ref)
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

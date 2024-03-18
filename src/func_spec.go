package src

import (
	"gomakus/utils"
)

// FuncSingleInput represents multiple components of single input value
// e.g. x Point input argument can correspond to a pair of selectors: VarSelector{VarId: 0, Selector: {"x"}}, VarSelector{VarId: 0, Selector: {"y"}}
type FuncSingleInput []VarSelector

// FuncMultiInput represents all input arguments of the function
type FuncMultiInput []FuncSingleInput

type FuncInputRef struct {
	ArgIndex      int
	SelectorIndex int
}

// FuncOutputRef represents relation between input & output parameters
// If no relation were found - InputRef.ArgIndex will be BlankVarId
// If generation of variable were changed - then GenChange will be +1/-1
type FuncOutputRef struct {
	InputRef   FuncInputRef
	OutputPath Path
	GenChange  GenChangeType
}

// FuncSingleOutput represents multiple components returned in single value (e.g.: return Point{x: 1, y: 2})
type FuncSingleOutput []FuncOutputRef

// FuncMultiOutput represents all return values of the function for one particular outcome
type FuncMultiOutput []FuncSingleOutput

// FuncSpec is a collection of all potential function outcomes (considering non-linear function control)
type FuncSpec struct {
	Inputs  FuncMultiInput
	Outputs FuncMultiOutput
}

type FuncSpecCollection map[FuncId]FuncSpec

func NewFuncSpec(inputs FuncMultiInput, outputs FuncMultiOutput) FuncSpec {
	inputSelectors := make(map[string]struct{})
	for _, input := range inputs {
		for _, selector := range input {
			selectorString := selector.String()
			_, ok := inputSelectors[selectorString]
			utils.Assertf(!ok, "input selectors should be unique for func spec: inputs=%#v, outputs=%#v", inputs, outputs)

			inputSelectors[selectorString] = struct{}{}
		}
	}
	for _, output := range outputs {
		for _, selector := range output {
			utils.Assertf(
				selector.InputRef.ArgIndex == BlankVarId ||
					(0 <= selector.InputRef.ArgIndex && selector.InputRef.ArgIndex < len(inputs) && 0 <= selector.InputRef.SelectorIndex && selector.InputRef.SelectorIndex < len(inputs[selector.InputRef.ArgIndex])),
				"input ref should point to defined input argument or BlankVarId: %#v", selector.InputRef,
			)
		}
	}
	return FuncSpec{Inputs: inputs, Outputs: outputs}
}

var (
	SliceFuncSpec = NewFuncSpec(
		FuncMultiInput{{{VarId: 0}}},
		FuncMultiOutput{{{InputRef: FuncInputRef{ArgIndex: 0}, GenChange: PrevGen}}},
	)
	AppendFuncSpec = NewFuncSpec(
		FuncMultiInput{{{VarId: 0}}},
		FuncMultiOutput{{{InputRef: FuncInputRef{ArgIndex: 0}, GenChange: NextGen}}},
	)

	DefaultFuncSpecCollection = map[FuncId]FuncSpec{
		SliceFuncId:  SliceFuncSpec,
		AppendFuncId: AppendFuncSpec,
	}
)

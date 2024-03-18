package src

// FuncSingleInput represents multiple components of single input value
// e.g. x Point input argument can correspond to a pair of selectors: VarSelector{VarId: 0, Selector: {"x"}}, VarSelector{VarId: 0, Selector: {"y"}}
type FuncSingleInput []VarSelector

// FuncMultiInput represents all input arguments of the function
type FuncMultiInput []FuncSingleInput

// FuncOutputRef represents relation between input & output parameters
// If no relation were found - InputArgEmbed.VarSelector.VarId will be BlankVarId
// If generation of variable were changed - then GenChange will be +1/-1
type FuncOutputRef struct {
	InputArgEmbed VarEmbed
	GenChange     GenChangeType
}

// FuncSingleOutput represents multiple components returned in single value (e.g.: return Point{x: 1, y: 2})
type FuncSingleOutput []FuncOutputRef

// FuncMultiOutput represents all return values of the function for one particular outcome
type FuncMultiOutput []FuncSingleOutput

// FuncSpec is a collection of all potential function outcomes (considering non-linear function control)
type FuncSpec struct {
	Inputs  []FuncMultiInput
	Outputs []FuncMultiOutput
}

var (
	SliceFuncSpec = FuncSpec{
		Inputs:  []FuncMultiInput{{{{VarId: 0}}}},
		Outputs: []FuncMultiOutput{{{{InputArgEmbed: VarEmbed{VarSelector: VarSelector{VarId: 0}}, GenChange: PrevGen}}}},
	}
	AppendFuncSpec = FuncSpec{
		Inputs:  []FuncMultiInput{{{{VarId: 0}}}},
		Outputs: []FuncMultiOutput{{{{InputArgEmbed: VarEmbed{VarSelector: VarSelector{VarId: 0}}, GenChange: NextGen}}}},
	}
)

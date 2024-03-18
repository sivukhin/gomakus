package src

type (
	FuncId int
	Path   []string
	// VarSelector represents field access for given var: v.x.y.field
	VarSelector struct {
		VarId    VarId
		Selector Path
	}
	// VarEmbed represents embedding of VarSelector at some path: { name: { value: v.x.y.field } }
	VarEmbed struct {
		Path        Path
		VarSelector VarSelector
	}
	// VarComposition represents collection of embedded VarSelector
	VarComposition []VarEmbed
)

func (e VarComposition) Embed(name string) VarComposition {
	result := make(VarComposition, len(e))
	for i, embed := range e {
		path := append([]string{name}, embed.Path...)
		result[i] = VarEmbed{Path: path, VarSelector: embed.VarSelector}
	}
	return result
}

func (p VarEmbed) Select(name string) (VarEmbed, bool) {
	if len(p.Path) > 0 && p.Path[0] == name {
		result := VarEmbed{Path: append([]string{}, p.Path[1:]...), VarSelector: p.VarSelector}
		return result, true
	} else if len(p.Path) == 0 && p.VarSelector.VarId != BlankVarId {
		result := VarEmbed{VarSelector: VarSelector{
			VarId:    p.VarSelector.VarId,
			Selector: append(append([]string(nil), p.VarSelector.Selector...), name),
		}}
		return result, true
	}
	return VarEmbed{}, false
}

// BlankVarId is a special variable which represent universal sink & source of data
// - If v1 = BlankVarId - then it behaves like a make() operation (so the previous value of v1 always reset with a fresh new instance)
// - If BlankVarId = v1 - then it behaves like a /dev/null and just consumes the data
const (
	BlankVarId          = -1
	BlankVarName string = "_"
)

// simple intermediate representation (IR) operations which we derive from the AST
type (
	Operation any
	// AssignSelectorOp "complex" operation: var1.a.b.c = var2.d.e
	AssignSelectorOp struct{ FromSelector, ToSelector VarSelector }
	// UseSelectorsOp "complex" operation: o1, o2, ... = f(i1, i2.b, {a: i3.c, b: i4.d.e}, ...)
	UseSelectorsOp struct {
		FuncId  FuncId
		Inputs  []VarComposition
		Outputs []VarId
	}
	// ReturnVarsOp "complex" operation: return v1, v2, ...
	ReturnVarsOp struct {
		VarIds []VarId
	}
	// NoOp artificial operation which simplify execution graph construction
	NoOp struct{}
	// AssignVarOp "primitive" operation: var1 = var2
	AssignVarOp struct{ FromVarId, ToVarId VarId }
	// ChangeVarGenOp "primitive" operation: var = next(var) / prev(var)
	ChangeVarGenOp struct {
		VarId     VarId
		GenChange int
	}
)

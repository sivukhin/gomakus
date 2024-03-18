package src

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFactorizeAssignments(t *testing.T) {
	// b = a.x
	// c = b
	// d = c.y
	// e = a.y
	rules := FactorizeAssignments([]AssignSelectorOp{
		{
			FromSelector: VarSelector{VarId: 0, Selector: Path{"x"}},
			ToSelector:   VarSelector{VarId: 1},
		},
		{
			FromSelector: VarSelector{VarId: 1},
			ToSelector:   VarSelector{VarId: 2},
		},
		{
			FromSelector: VarSelector{VarId: 2, Selector: Path{"y"}},
			ToSelector:   VarSelector{VarId: 3},
		},
		{
			FromSelector: VarSelector{VarId: 0, Selector: Path{"y"}},
			ToSelector:   VarSelector{VarId: 4},
		},
	})
	t.Log(rules)
	require.Equal(t, FactorizationRules{
		0: {{"x", "y"}, {"y"}},
		1: {{"y"}},
		2: {{"y"}},
		3: {{}},
		4: {{}},
	}, rules)
}

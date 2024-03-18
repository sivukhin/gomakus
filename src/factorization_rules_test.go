package src

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFactorizationRules(t *testing.T) {
	rules := FactorizationRules{
		0: []Path{{"x", "a"}, {"x", "b"}, {"z"}},
		1: []Path{{"a"}, {"b"}},
	}
	factorized := rules.FactorizeSelector(VarSelector{VarId: 0, Selector: Path{"x"}})
	require.Equal(t, []VarSelector{{VarId: 0, Selector: Path{"x", "a"}}, {VarId: 0, Selector: Path{"x", "b"}}}, factorized)
}

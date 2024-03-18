package src

import (
	"slices"
	"sort"

	"go-append-check/utils"
)

// FactorizationRules stores all useful paths for every variable in single function
// For example, {a: [x.y, z], b: [y], c: [y], d: [], e: []} is the factorization rules for the following snippet:
// b = a.x
// c = b
// d = c.y
// e = a.y
type FactorizationRules map[VarId][]Path

func (r FactorizationRules) FactorizeSelector(selector VarSelector) []VarSelector {
	paths, ok := r[selector.VarId]
	if !ok {
		return []VarSelector{selector}
	}
	factorized := make([]VarSelector, 0)
	for _, factorizedPath := range paths {
		if hasPrefix(factorizedPath, selector.Selector) {
			factorized = append(factorized, VarSelector{
				VarId:    selector.VarId,
				Selector: factorizedPath,
			})
		}
	}
	utils.Assertf(len(factorized) > 0, "factorization rules has no mention for selector %+v", selector)
	// we need to sort factorized path in order for them to match as an operands of assignment operator
	sort.Slice(factorized, func(a, b int) bool { return slices.Compare(factorized[a].Selector, factorized[b].Selector) < 0 })
	return factorized
}

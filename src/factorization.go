package src

import (
	"slices"
	"sort"

	"gomakus/utils"
)

type factorizationEntry struct {
	varId VarId
	path  string
}

type factorizationContext struct {
	analyzed  map[factorizationEntry]struct{}
	rules     FactorizationRules
	sources   map[VarId][]AssignSelectorOp
	targets   map[VarId][]AssignSelectorOp
	workspace []byte
}

func FactorizeAssignments(assigns []AssignSelectorOp) FactorizationRules {
	targets := make(map[VarId][]AssignSelectorOp)
	sources := make(map[VarId][]AssignSelectorOp)
	for _, assign := range assigns {
		if assign.FromSelector.VarId == BlankVarId || assign.ToSelector.VarId == BlankVarId {
			continue
		}
		// todo (sivukhin, 2024-03-23): simple hack to forbid trivial cycles
		if assign.FromSelector.VarId == assign.ToSelector.VarId {
			continue
		}
		targets[assign.FromSelector.VarId] = append(targets[assign.FromSelector.VarId], assign)
		sources[assign.ToSelector.VarId] = append(sources[assign.ToSelector.VarId], assign)
	}

	var workspace [1024]byte
	context := factorizationContext{
		analyzed:  make(map[factorizationEntry]struct{}),
		rules:     make(FactorizationRules),
		sources:   sources,
		targets:   targets,
		workspace: workspace[:0],
	}
	for _, assign := range assigns {
		factorizeAssigment(context, assign.FromSelector.VarId, Path{})
		factorizeAssigment(context, assign.FromSelector.VarId, assign.FromSelector.Selector)
		factorizeAssigment(context, assign.ToSelector.VarId, Path{})
		factorizeAssigment(context, assign.ToSelector.VarId, assign.ToSelector.Selector)
	}
	for varId, paths := range context.rules {
		sort.Slice(paths, func(i, j int) bool { return slices.Compare(paths[i], paths[j]) < 0 })
		leafs := make([]Path, 0)
		for i := 1; i <= len(paths); i++ {
			if i == len(paths) || !hasPrefix(paths[i], paths[i-1]) {
				leafs = append(leafs, paths[i-1])
			}
		}
		context.rules[varId] = leafs
	}
	return context.rules
}

func factorizeAssigment(context factorizationContext, varId VarId, path Path) {
	if varId == BlankVarId {
		return
	}
	pathBytes := joinTo(context.workspace, path, ",")
	entry := factorizationEntry{varId: varId, path: string(pathBytes)}
	if _, ok := context.analyzed[entry]; ok {
		return
	}
	context.rules[varId] = append(context.rules[varId], path)
	context.analyzed[entry] = struct{}{}

	for _, source := range context.sources[varId] {
		if !hasPrefix(path, source.ToSelector.Selector) {
			continue
		}
		targetPath := append(append([]string{}, source.FromSelector.Selector...), path[len(source.ToSelector.Selector):]...)
		factorizeAssigment(context, source.FromSelector.VarId, targetPath)
	}
	for _, target := range context.targets[varId] {
		if !hasPrefix(path, target.FromSelector.Selector) {
			continue
		}
		targetPath := append(append([]string{}, target.ToSelector.Selector...), path[len(target.FromSelector.Selector):]...)
		factorizeAssigment(context, target.ToSelector.VarId, targetPath)
	}
}

func hasPrefix[T comparable](haystack, needle []T) bool {
	if len(haystack) < len(needle) {
		return false
	}
	for i := range needle {
		if needle[i] != haystack[i] {
			return false
		}
	}
	return true
}

func joinTo(workspace []byte, elements []string, separator string) []byte {
	utils.Assertf(len(workspace) == 0, "workspace should be empty")
	utils.Assertf(cap(workspace) > 0, "workspace should have non-zero capacity")
	for i, element := range elements {
		if i > 0 {
			workspace = append(workspace, []byte(separator)...)
		}
		workspace = append(workspace, []byte(element)...)
	}
	return workspace
}

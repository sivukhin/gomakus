package src

import (
	"slices"
)

func TopologyOrder[T comparable](root T, next func(v T) []T) []T {
	var order []T
	topologyOrder(root, next, make(map[T]struct{}), &order)
	slices.Reverse(order)
	return order
}

func topologyOrder[T comparable](v T, next func(v T) []T, visited map[T]struct{}, order *[]T) {
	if _, ok := visited[v]; ok {
		return
	}
	visited[v] = struct{}{}
	for _, u := range next(v) {
		topologyOrder(u, next, visited, order)
	}
	*order = append(*order, v)
}

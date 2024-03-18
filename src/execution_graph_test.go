package src

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateTrace(t *testing.T) {
	t.Run("a=new();b=a;a=next(a);b=next(b)", func(t *testing.T) {
		trace := ExecutionTrace{
			{Id: 0, Statement: AssignmentStatement{Target: 1, Source: 0, GenChange: SameGen}},
			{Id: 1, Statement: AssignmentStatement{Target: 2, Source: 1, GenChange: SameGen}},
			{Id: 2, Statement: AssignmentStatement{Target: 1, Source: 1, GenChange: NextGen}},
			{Id: 3, Statement: AssignmentStatement{Target: 2, Source: 2, GenChange: NextGen}},
		}
		validation := ExecutionContext{}.ValidateTrace(trace)
		require.Equal(t, []ExecutionError{{Trace: trace, Id: 3, ConflictOrigin: 0, ConflictVariable: 2}}, validation)
	})
	t.Run("a=new();a=next(a);a=new();a=next()", func(t *testing.T) {
		validation := ExecutionContext{}.ValidateTrace(ExecutionTrace{
			{Id: 0, Statement: AssignmentStatement{Target: 2, Source: 0, GenChange: SameGen}},
			{Id: 1, Statement: AssignmentStatement{Target: 2, Source: 2, GenChange: NextGen}},
			{Id: 2, Statement: AssignmentStatement{Target: 2, Source: 1, GenChange: SameGen}},
			{Id: 3, Statement: AssignmentStatement{Target: 2, Source: 2, GenChange: NextGen}},
		})
		require.Empty(t, validation)
	})
	t.Run("a=new();b=a;a=next(a);b=new();b=next(b)", func(t *testing.T) {
		validation := ExecutionContext{}.ValidateTrace(ExecutionTrace{
			{Id: 0, Statement: AssignmentStatement{Target: 2, Source: 0, GenChange: SameGen}},
			{Id: 1, Statement: AssignmentStatement{Target: 3, Source: 2, GenChange: SameGen}},
			{Id: 2, Statement: AssignmentStatement{Target: 2, Source: 2, GenChange: NextGen}},
			{Id: 3, Statement: AssignmentStatement{Target: 3, Source: 1, GenChange: SameGen}},
			{Id: 3, Statement: AssignmentStatement{Target: 3, Source: 3, GenChange: NextGen}},
		})
		require.Empty(t, validation)
	})
	t.Run("a=new();b=a;a=next(a);call(a)", func(t *testing.T) {
		ctx := ExecutionContext{NextGenArguments: map[StatementId][]int{3: {0}}}
		validation := ctx.ValidateTrace(ExecutionTrace{
			{Id: 0, Statement: AssignmentStatement{Target: 1, Source: 0, GenChange: SameGen}},
			{Id: 1, Statement: AssignmentStatement{Target: 2, Source: 1, GenChange: SameGen}},
			{Id: 2, Statement: AssignmentStatement{Target: 1, Source: 1, GenChange: NextGen}},
			{Id: 3, Statement: UseStatement{Values: []VarId{1}}},
		})
		require.Empty(t, validation)
	})
	t.Run("a=new();b=a;a=next(a);call(b)", func(t *testing.T) {
		ctx := ExecutionContext{NextGenArguments: map[StatementId][]int{3: {0}}}
		trace := ExecutionTrace{
			{Id: 0, Statement: AssignmentStatement{Target: 1, Source: 0, GenChange: SameGen}},
			{Id: 1, Statement: AssignmentStatement{Target: 2, Source: 1, GenChange: SameGen}},
			{Id: 2, Statement: AssignmentStatement{Target: 1, Source: 1, GenChange: NextGen}},
			{Id: 3, Statement: UseStatement{Values: []VarId{2}}},
		}
		validation := ctx.ValidateTrace(trace)
		require.Equal(t, []ExecutionError{{Trace: trace, Id: 3, ConflictOrigin: 0, ConflictVariable: 2}}, validation)
	})
	t.Run("a=new();b=prev(a);call(b)", func(t *testing.T) {
		ctx := ExecutionContext{NextGenArguments: map[StatementId][]int{2: {0}}}
		trace := ExecutionTrace{
			{Id: 0, Statement: AssignmentStatement{Target: 1, Source: 0, GenChange: SameGen}},
			{Id: 1, Statement: AssignmentStatement{Target: 2, Source: 1, GenChange: PrevGen}},
			{Id: 2, Statement: UseStatement{Values: []VarId{2}}},
		}
		validation := ctx.ValidateTrace(trace)
		require.Equal(t, []ExecutionError{{Trace: trace, Id: 2, ConflictOrigin: 0, ConflictVariable: 2}}, validation)
	})
}

func TestValidateGraph(t *testing.T) {
	t.Run("a=new();b=a;a=next(a);[{c=b},{c=next(b)}]", func(t *testing.T) {
		graph := NewExecutionGraph()
		s0 := AssignmentStatement{Target: 1, Source: 0, GenChange: SameGen}
		s1 := AssignmentStatement{Target: 2, Source: 1, GenChange: SameGen}
		s2 := AssignmentStatement{Target: 1, Source: 1, GenChange: NextGen}
		s3 := AssignmentStatement{Target: 3, Source: 2, GenChange: SameGen}
		s4 := AssignmentStatement{Target: 3, Source: 2, GenChange: NextGen}
		n0 := graph.NewNode(s0)
		n1 := graph.NewNode(s1)
		n2 := graph.NewNode(s2)
		n3 := graph.NewNode(s3)
		n4 := graph.NewNode(s4)
		graph.AddEdge(n0, n1)
		graph.AddEdge(n1, n2)
		graph.AddEdge(n2, n3)
		graph.AddEdge(n2, n4)

		validation := ExecutionContext{}.ValidateGraph(graph, 0, 1)
		trace := ExecutionTrace{
			{Id: 0, Statement: s0},
			{Id: 1, Statement: s1},
			{Id: 2, Statement: s2},
			{Id: 4, Statement: s4},
		}
		require.Equal(t, []ExecutionError{{Trace: trace, Id: 4, ConflictOrigin: 0, ConflictVariable: 2}}, validation)
	})
	t.Run("a=new();loop[a=next(a);]", func(t *testing.T) {
		graph := NewExecutionGraph()
		s0 := AssignmentStatement{Target: 1, Source: 0, GenChange: SameGen}
		s1 := AssignmentStatement{Target: 1, Source: 1, GenChange: NextGen}
		n0 := graph.NewNode(s0)
		n1 := graph.NewNode(s1)
		graph.AddEdge(n0, n1)
		graph.AddEdge(n1, n1)

		validation := ExecutionContext{}.ValidateGraph(graph, 0, 2)
		require.Empty(t, validation)
	})
	t.Run("a=new();loop[a=next(a);]", func(t *testing.T) {
		graph := NewExecutionGraph()
		s0 := AssignmentStatement{Target: 0, Source: 1, GenChange: SameGen}
		s1 := AssignmentStatement{Target: 2, Source: 0, GenChange: NextGen}
		s2 := AssignmentStatement{Target: 0, Source: 2, GenChange: SameGen}
		n0 := graph.NewNode(s0)
		n1 := graph.NewNode(s1)
		n2 := graph.NewNode(s2)
		graph.AddEdge(n0, n1)
		graph.AddEdge(n1, n2)
		graph.AddEdge(n2, n1)

		validation := ExecutionContext{}.ValidateGraph(graph, 0, 2)
		require.Empty(t, validation)
	})
}

func TestGenerateTraces(t *testing.T) {
	t.Run("a->b->c->d", func(t *testing.T) {
		graph := NewExecutionGraph()
		s0 := AssignmentStatement{Target: 1, Source: 0, GenChange: SameGen}
		s1 := AssignmentStatement{Target: 2, Source: 1, GenChange: SameGen}
		s2 := AssignmentStatement{Target: 1, Source: 1, GenChange: NextGen}
		s3 := UseStatement{Values: []VarId{1}}
		n0 := graph.NewNode(s0)
		n1 := graph.NewNode(s1)
		n2 := graph.NewNode(s2)
		n3 := graph.NewNode(s3)
		graph.AddEdge(n0, n1)
		graph.AddEdge(n1, n2)
		graph.AddEdge(n2, n3)

		traces := graph.GenerateTraces(0, 1)
		require.Equal(t, []ExecutionTrace{
			{
				{Id: 0, Statement: s0},
				{Id: 1, Statement: s1},
				{Id: 2, Statement: s2},
				{Id: 3, Statement: s3},
			},
		}, traces)
	})
	t.Run("a->b->[c,d]", func(t *testing.T) {
		graph := NewExecutionGraph()
		s0 := AssignmentStatement{Target: 1, Source: 0, GenChange: SameGen}
		s1 := AssignmentStatement{Target: 2, Source: 0, GenChange: SameGen}
		s2 := AssignmentStatement{Target: 3, Source: 1, GenChange: NextGen}
		s3 := AssignmentStatement{Target: 3, Source: 2, GenChange: NextGen}
		n0 := graph.NewNode(s0)
		n1 := graph.NewNode(s1)
		n2 := graph.NewNode(s2)
		n3 := graph.NewNode(s3)
		graph.AddEdge(n0, n1)
		graph.AddEdge(n1, n2)
		graph.AddEdge(n1, n3)

		traces := graph.GenerateTraces(0, 1)
		require.Equal(t, []ExecutionTrace{
			{
				{Id: 0, Statement: s0},
				{Id: 1, Statement: s1},
				{Id: 2, Statement: s2},
			},
			{
				{Id: 0, Statement: s0},
				{Id: 1, Statement: s1},
				{Id: 3, Statement: s3},
			},
		}, traces)
	})
	t.Run("a->loop[b]/repeatLimit=1", func(t *testing.T) {
		graph := NewExecutionGraph()
		s0 := AssignmentStatement{Target: 1, Source: 0, GenChange: SameGen}
		s1 := AssignmentStatement{Target: 2, Source: 1, GenChange: NextGen}
		n0 := graph.NewNode(s0)
		n1 := graph.NewNode(s1)
		graph.AddEdge(n0, n1)
		graph.AddEdge(n1, n1)
		require.Equal(t, []ExecutionTrace{
			{
				{Id: 0, Statement: s0},
				{Id: 1, Statement: s1},
			},
		}, graph.GenerateTraces(0, 1))
		require.Equal(t, []ExecutionTrace{
			{
				{Id: 0, Statement: s0},
				{Id: 1, Statement: s1},
				{Id: 1, Statement: s1},
			},
		}, graph.GenerateTraces(0, 2))
	})
}

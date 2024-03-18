package main

import (
	"fmt"
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/singlechecker"

	"go-append-check/src"
)

var Analyzer = &analysis.Analyzer{
	Name: "go_append_check",
	Doc:  "detect potential overwrites of slice elements with append functions",
	Run:  run,
}

func buildExpressionGraph(
	fset *token.FileSet,
	scopes src.Scopes,
	cursor src.ExecutionGraphCursor,
	someExpression ast.Expr,
) (src.VarId, src.ExecutionGraphCursor, error) {
	switch expression := someExpression.(type) {
	case *ast.Ident:
		return scopes.GetVar(expression.Name), cursor, nil
	case *ast.SliceExpr:
		variable, nextCursor, err := buildExpressionGraph(fset, scopes, cursor, expression.X)
		if err != nil {
			return 0, src.ExecutionGraphCursor{}, err
		}
		sliceVariable := scopes.NewVarId()
		genChange := src.PrevGen
		if expression.Max == nil {
			genChange = src.SameGen
		}
		nextCursor.Next(src.AssignmentStatement{Target: sliceVariable, Source: variable, GenChange: genChange})
		return sliceVariable, nextCursor, nil
	case *ast.CallExpr:
		if ident, ok := expression.Fun.(*ast.Ident); ok && ident.Name == "append" {
			variable, nextCursor, err := buildExpressionGraph(fset, scopes, cursor, expression.Args[0])
			if err != nil {
				return 0, src.ExecutionGraphCursor{}, err
			}
			appendVariable := scopes.NewVarId()
			nextCursor.Next(src.AssignmentStatement{Target: appendVariable, Source: variable, GenChange: src.NextGen})
			return appendVariable, nextCursor, nil
		}
		return 0, src.ExecutionGraphCursor{}, fmt.Errorf("unable to process expression")
	case *ast.SelectorExpr, // todo: handler selector correctly
		*ast.BadExpr,
		*ast.Ellipsis,
		*ast.BasicLit,
		*ast.FuncLit,
		*ast.CompositeLit,
		*ast.ParenExpr,
		*ast.IndexExpr,
		*ast.IndexListExpr,
		*ast.TypeAssertExpr,
		*ast.StarExpr,
		*ast.UnaryExpr,
		*ast.BinaryExpr,
		*ast.KeyValueExpr,
		*ast.ArrayType,
		*ast.StructType,
		*ast.FuncType,
		*ast.InterfaceType,
		*ast.MapType,
		*ast.ChanType:
		return 0, src.ExecutionGraphCursor{}, fmt.Errorf("unable to process expression")
	}
	panic(fmt.Errorf("unexpected expression: %#v", someExpression))
}

func test() {
	candidatePostsToLog := make([]int, 0)
	for i := 0; i < 1; i++ {
		var propsToLog []int
		if true {
			propsToLog = append(candidatePostsToLog, 1)
		}
		candidatePostsToLog = append(propsToLog, 2)
	}
}

func buildStatementsGraph(
	fset *token.FileSet,
	scopes src.Scopes,
	cursor src.ExecutionGraphCursor,
	statements []ast.Stmt,
) (src.ExecutionGraphCursor, error) {
	for _, someStatement := range statements {
		for {
			if label, ok := someStatement.(*ast.LabeledStmt); ok {
				someStatement = label.Stmt
			} else {
				break
			}
		}

		switch statement := someStatement.(type) {
		case *ast.BlockStmt:
			var err error
			cursor, err = buildStatementsGraph(fset, scopes.PushScope(), cursor, statement.List)
			if err != nil {
				return src.ExecutionGraphCursor{}, err
			}
		case *ast.IfStmt:
			chain := statement
			finish := src.NewExecutionGraphCursor(cursor.Graph)
			for chain != nil {
				end, err := buildStatementsGraph(fset, scopes.PushScope(), cursor, chain.Body.List)
				if err != nil {
					return src.ExecutionGraphCursor{}, err
				}
				finish.Graph.AddEdge(end.Current, finish.Current)
				if block, ok := chain.Else.(*ast.BlockStmt); ok {
					chain = &ast.IfStmt{Body: block}
				} else if nested, ok := chain.Else.(*ast.IfStmt); ok {
					chain = nested
				} else if chain.Else == nil {
					break
				} else {
					panic(fmt.Errorf("unexpected if/else structure: %#v", fset.Position(statement.Pos())))
				}
			}
			cursor = finish
		case *ast.ForStmt:
			cursor.Next(src.NopStatement{})
			end, err := buildStatementsGraph(fset, scopes.PushScope(), cursor, statement.Body.List)
			if err != nil {
				return src.ExecutionGraphCursor{}, err
			}
			cursor.Graph.AddEdge(end.Current, cursor.Current)
			cursor = end
		case *ast.RangeStmt:
			cursor.Next(src.NopStatement{})
			end, err := buildStatementsGraph(fset, scopes.PushScope(), cursor, statement.Body.List)
			if err != nil {
				return src.ExecutionGraphCursor{}, err
			}
			cursor.Graph.AddEdge(end.Current, cursor.Current)
			cursor = end
		case *ast.AssignStmt:
			for i := range statement.Rhs {
				lhs, rhs := statement.Lhs[i], statement.Rhs[i]
				var leftVar, rightVar src.VarId
				leftVar, lCursor, err := buildExpressionGraph(fset, scopes, cursor, lhs)
				if err != nil {
					continue
				}
				cursor = lCursor
				rightVar, rCursor, err := buildExpressionGraph(fset, scopes, cursor, rhs)
				if err != nil {
					continue
				}
				cursor = rCursor
				cursor.Next(src.AssignmentStatement{Target: leftVar, Source: rightVar, GenChange: src.SameGen})
			}
		case *ast.DeclStmt:
			switch decl := statement.Decl.(type) {
			case *ast.GenDecl:
				for _, spec := range decl.Specs {
					valueSpec, ok := spec.(*ast.ValueSpec)
					if !ok {
						continue
					}
					for _, name := range valueSpec.Names {
						_ = scopes.GetVar(name.Name)
					}
				}
			}
		case *ast.ExprStmt,
			*ast.BranchStmt,
			*ast.DeferStmt,
			*ast.EmptyStmt,
			*ast.GoStmt,
			*ast.IncDecStmt,
			*ast.ReturnStmt,
			*ast.SelectStmt,
			*ast.SendStmt,
			*ast.TypeSwitchStmt,
			*ast.SwitchStmt:
			continue
		case *ast.BadStmt:
			return src.ExecutionGraphCursor{}, fmt.Errorf("BadStmt encountered: %#v", statement)
		case *ast.LabeledStmt:
			panic(fmt.Errorf("LabeledStmt encountered"))
		default:
			panic(fmt.Errorf("unexpected statement: %#v", statement))
		}
	}
	return cursor, nil
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			funcDecl, ok := node.(*ast.FuncDecl)
			if !ok {
				return true
			}
			if funcDecl.Body == nil {
				return true
			}
			graph := src.NewExecutionGraph()
			cursor := src.NewExecutionGraphCursor(graph)
			//if funcDecl.Name.Name == "test" {
			_, _ = buildStatementsGraph(pass.Fset, src.NewScopes(), cursor, funcDecl.Body.List)
			validationErrs := src.ExecutionContext{}.ValidateGraph(graph, cursor.Current, 2)
			if len(validationErrs) != 0 {
				pass.Reportf(funcDecl.Pos(), "potential append overwrite found in function %v", funcDecl.Name)
			}
			//}
			return true
		})
	}
	return nil, nil
}

func main() {
	singlechecker.Main(Analyzer)
}

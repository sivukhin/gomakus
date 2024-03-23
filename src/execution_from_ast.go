package src

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/sivukhin/gomakus/utils"
)

// deconstructIf analyze chain of if / else if / else blocks and return flat list representing single chain (without analyzing nested conditions)
func deconstructIf(fset *token.FileSet, ifStmt *ast.IfStmt) (inits []ast.Stmt, bodies []ast.Stmt) {
	inits, bodies = make([]ast.Stmt, 0), make([]ast.Stmt, 0)
	for ifStmt != nil {
		inits = append(inits, ifStmt.Init)
		bodies = append(bodies, ifStmt.Body)

		if block, ok := ifStmt.Else.(*ast.BlockStmt); ok {
			ifStmt = &ast.IfStmt{Body: block}
		} else if nested, ok := ifStmt.Else.(*ast.IfStmt); ok {
			ifStmt = nested
		} else if ifStmt.Else == nil {
			break
		} else {
			panic(fmt.Errorf("unexpected if/else structure: %#v", fset.Position(ifStmt.Pos())))
		}
	}
	utils.Assertf(len(inits) == len(bodies), "deconstructIf invariant failed at %v", fset.Position(ifStmt.Pos()))
	return
}

// deconstructDecl recognizes two declaration methods: var x string = "s" and x := "s"
func deconstructDecl(stmt ast.Stmt) ([]string, []ast.Expr, bool) {
	names, values := make([]string, 0), make([]ast.Expr, 0)
	if decl, ok := stmt.(*ast.DeclStmt); ok {
		genDecl := decl.Decl.(*ast.GenDecl)
		if genDecl.Tok != token.VAR { // ast.GenDecl also includes const def, imports and type def
			return nil, nil, false
		}
		for _, spec := range genDecl.Specs {
			valueSpec := spec.(*ast.ValueSpec) // under var declaration we must encounter only ValueSpecs
			for i, name := range valueSpec.Names {
				names = append(names, name.Name)
				if len(valueSpec.Values) == 0 { // e.g. var x, y, z string
					values = append(values, nil)
				} else {
					values = append(values, valueSpec.Values[i])
				}
			}
		}
		return names, values, true
	}
	if assign, ok := stmt.(*ast.AssignStmt); ok && assign.Tok == token.DEFINE {
		for _, lhs := range assign.Lhs {
			names = append(names, lhs.(*ast.Ident).Name)
		}
		for _, rhs := range assign.Rhs {
			values = append(values, rhs)
		}
		return names, values, true
	}
	return nil, nil, false
}

func executionFromVarComposition(
	fset *token.FileSet,
	node ast.Node,
	builder ExecutionBuilder,
	varSelector VarSelector,
	varCompositions VarComposition,
) ExecutionBuilder {
	if len(varCompositions) == 0 {
		builder = builder.ApplyNextWithRef(AssignSelectorOp{
			FromSelector: VarSelector{VarId: BlankVarId},
			ToSelector:   varSelector,
		}, node.Pos())
	}
	for _, assignOp := range varCompositions {
		builder = builder.ApplyNextWithRef(AssignSelectorOp{
			FromSelector: assignOp.VarSelector,
			ToSelector:   VarSelector{VarId: varSelector.VarId, Selector: append(varSelector.Selector, assignOp.Path...)},
		}, node.Pos())
	}
	return builder
}

func executionFromExpr(
	builder ExecutionBuilder,
	scopes Scopes,
	fset *token.FileSet,
	expr ast.Expr,
	exprOutputs int,
) (ExecutionBuilder, []VarComposition) {
	blanks := make([]VarComposition, exprOutputs)
	for i := range blanks {
		blanks[i] = VarComposition{{VarSelector: VarSelector{VarId: BlankVarId}}}
	}

	switch e := expr.(type) {
	case *ast.ParenExpr:
		return executionFromExpr(builder, scopes, fset, e.X, exprOutputs)
	case *ast.Ident:
		utils.Assertf(exprOutputs == 1, "unexpected multi-output expression: %v", fset.Position(expr.Pos()))
		return builder, []VarComposition{{{VarSelector: VarSelector{VarId: scopes.GetVarOrBlank(e.Name)}}}}
	case *ast.SelectorExpr:
		var varCompositions []VarComposition
		builder, varCompositions = executionFromExpr(builder, scopes, fset, e.X, 1)
		utils.Assertf(len(varCompositions) == 1, "selector must be applied to single value")
		for _, varEmbed := range varCompositions[0] {
			if selected, ok := varEmbed.Select(e.Sel.Name); ok {
				return builder, []VarComposition{{selected}}
			}
		}
		return builder, blanks
	case *ast.CompositeLit:
		return builder, blanks
		//if _, ok := e.Type.(*ast.ArrayType); ok {
		//	return builder, blanks
		//}
		//var varCompositions VarComposition
		//for _, el := range e.Elts {
		//	keyValueExp, ok := el.(*ast.KeyValueExpr)
		//	if !ok {
		//		continue
		//	}
		//	keyIdent, ok := keyValueExp.Key.(*ast.Ident)
		//	if !ok {
		//		continue
		//	}
		//	var nestedVarComposition []VarComposition
		//	builder, nestedVarComposition = executionFromExpr(builder, scopes, fset, keyValueExp.Value, 1)
		//	utils.Assertf(len(nestedVarComposition) == 1, "composite lit key must be assigned to single value")
		//	varCompositions = append(varCompositions, nestedVarComposition[0].Embed(keyIdent.Name)...)
		//}
		//return builder, []VarComposition{varCompositions}
	case *ast.CallExpr, *ast.SliceExpr:
		var funcId FuncId
		var args []ast.Expr
		if call, isCall := e.(*ast.CallExpr); isCall {
			funcIdent, ok := call.Fun.(*ast.Ident)
			if ok {
				funcId, ok = scopes.TryGetFunc(funcIdent.Name)
			}
			if !ok {
				return builder, blanks
			}
			args = call.Args
		} else {
			slice := e.(*ast.SliceExpr)
			if slice.Max == nil {
				return executionFromExpr(builder, scopes, fset, slice.X, 1)
			}
			funcId = scopes.GetFunc(SliceFuncName)
			args = []ast.Expr{slice.X}
		}
		var inVarCompositions []VarComposition
		var outVarCompositions []VarComposition
		var outVars []VarId
		for _, arg := range args {
			var argVarComposition []VarComposition
			builder, argVarComposition = executionFromExpr(builder, scopes, fset, arg, 1)
			utils.Assertf(len(argVarComposition) == 1, "argument expression must have single value: %v", fset.Position(e.Pos()))
			inVarCompositions = append(inVarCompositions, argVarComposition[0])
		}
		for i := 0; i < exprOutputs; i++ {
			varId := scopes.NewVarId()
			outVars = append(outVars, varId)
			outVarCompositions = append(outVarCompositions, VarComposition{{VarSelector: VarSelector{VarId: varId}}})
		}
		builder = builder.ApplyNextWithRef(UseSelectorsOp{
			FuncId:  funcId,
			Inputs:  inVarCompositions,
			Outputs: outVars,
		}, expr.Pos())
		return builder, outVarCompositions
	case
		nil,
		*ast.StructType,
		*ast.Ellipsis,
		*ast.BasicLit,
		*ast.FuncLit,
		*ast.IndexExpr,
		*ast.IndexListExpr,
		*ast.TypeAssertExpr,
		*ast.StarExpr,
		*ast.UnaryExpr,
		*ast.BinaryExpr,
		*ast.KeyValueExpr,
		*ast.ArrayType,
		*ast.FuncType,
		*ast.InterfaceType,
		*ast.MapType,
		*ast.ChanType:
		return builder, blanks
	}
	panic(fmt.Errorf("unexpected expression"))
}

func executionFromStmtList(
	builder ExecutionBuilder,
	scopes Scopes,
	fset *token.FileSet,
	stmts []ast.Stmt,
	returnOutputs int,
) ExecutionBuilder {
	for _, s := range stmts {
		builder = executionFromStmt(builder, scopes, fset, s, returnOutputs)
	}
	return builder
}

func f() error {
	return nil
}

func executionFromStmt(
	builder ExecutionBuilder,
	scopes Scopes,
	fset *token.FileSet,
	stmt ast.Stmt,
	returnOutputs int,
) ExecutionBuilder {
	switch s := stmt.(type) {
	case *ast.LabeledStmt:
		return executionFromStmt(builder, scopes, fset, s.Stmt, returnOutputs)
	case *ast.BlockStmt:
		return executionFromStmtList(builder, scopes, fset, s.List, returnOutputs)
	case *ast.IfStmt:
		inits, bodies := deconstructIf(fset, s)
		scopes = scopes.PushScope() // all init stmts works in the separate scope, different from the external
		afterIf := builder.AcquirePointWithRef(s.End())
		for i, init := range inits {
			if init != nil {
				builder = executionFromStmt(builder, scopes, fset, init, returnOutputs)
			}
			executionFromStmt(builder, scopes.PushScope(), fset, bodies[i], returnOutputs).ConnectTo(afterIf.CurrentPoint)
		}
		return afterIf
	case *ast.ForStmt:
		scopes = scopes.PushScope()
		if s.Init != nil {
			builder = executionFromStmt(builder, scopes, fset, s.Init, returnOutputs)
		}
		beforeFor := builder

		builder = executionFromStmt(builder, scopes, fset, s.Body, returnOutputs)
		if s.Post != nil {
			builder = executionFromStmt(builder, scopes, fset, s.Post, returnOutputs)
		}
		builder, _ = executionFromExpr(builder, scopes, fset, s.Cond, 1)
		builder.ConnectTo(beforeFor.CurrentPoint)
		return builder
	case *ast.RangeStmt:
		scopes = scopes.PushScope()
		beforeFor := builder
		// create key, value in scope and reset them in IR
		if s.Key != nil {
			builder = builder.ApplyNextWithRef(AssignSelectorOp{
				ToSelector:   VarSelector{VarId: scopes.CreateVar(s.Key.(*ast.Ident).Name)},
				FromSelector: VarSelector{VarId: BlankVarId},
			}, s.Key.Pos())
		}
		if s.Value != nil {
			builder = builder.ApplyNextWithRef(AssignSelectorOp{
				ToSelector:   VarSelector{VarId: scopes.CreateVar(s.Value.(*ast.Ident).Name)},
				FromSelector: VarSelector{VarId: BlankVarId},
			}, s.Value.Pos())
		}
		builder = executionFromStmt(builder, scopes, fset, s.Body, returnOutputs)
		builder.ConnectTo(beforeFor.CurrentPoint)
		return builder
	case *ast.DeclStmt, *ast.AssignStmt:
		names, values, isDecl := deconstructDecl(stmt)
		if isDecl {
			utils.Assertf(len(values) == 1 || len(values) == len(names), "decl initial inputs/outputs count mismatch: %v", fset.Position(s.Pos()))
			valueOutputs := len(names)
			if len(values) == len(names) {
				valueOutputs = 1
			}

			var varCompositions []VarComposition
			for _, value := range values {
				var valueVarComposition []VarComposition
				builder, valueVarComposition = executionFromExpr(builder, scopes, fset, value, valueOutputs)
				utils.Assertf(len(valueVarComposition) == valueOutputs, "expression must have %v outputs: %v", valueOutputs, fset.Position(s.Pos()))
				varCompositions = append(varCompositions, valueVarComposition...)
			}
			for i, name := range names {
				varId := scopes.CreateVar(name)
				builder = executionFromVarComposition(fset, s, builder, VarSelector{VarId: varId}, varCompositions[i])
			}
		} else if assign, ok := stmt.(*ast.AssignStmt); ok {
			utils.Assertf(len(assign.Rhs) == 1 || len(assign.Lhs) == len(assign.Rhs), "assign initial inputs/outputs count mismatch: %v", fset.Position(s.Pos()))
			valueOutputs := len(assign.Lhs)
			if len(assign.Lhs) == len(assign.Rhs) {
				valueOutputs = 1
			}

			var varCompositions []VarComposition
			for _, value := range assign.Rhs {
				var valueVarComposition []VarComposition
				builder, valueVarComposition = executionFromExpr(builder, scopes, fset, value, valueOutputs)
				varCompositions = append(varCompositions, valueVarComposition...)
			}
			utils.Assertf(len(varCompositions) == len(assign.Lhs), "assign final inputs/outputs count mismatch: %v (%v != %v)", fset.Position(s.Pos()), len(varCompositions), len(assign.Lhs))
			for i := range assign.Lhs {
				var lhsVarComposition []VarComposition
				builder, lhsVarComposition = executionFromExpr(builder, scopes, fset, assign.Lhs[i], 1)
				utils.Assertf(len(lhsVarComposition) == 1, "lhs should have single value: %v", fset.Position(s.Pos()))
				// there can be more complex lhs which will be hard to analyze:
				// func f(a, b, c T) *T { return &b }
				// f(a, b, c).x.y = 1
				if len(lhsVarComposition[0]) != 1 {
					continue
				}
				if len(lhsVarComposition[0][0].Path) != 0 {
					continue
				}
				builder = executionFromVarComposition(fset, s, builder, lhsVarComposition[0][0].VarSelector, varCompositions[i])
			}
		}
		return builder
	case *ast.SwitchStmt:
		scopes = scopes.PushScope()
		builder = executionFromStmt(builder, scopes, fset, s.Init, returnOutputs)
		afterSwitch := builder.AcquirePointWithRef(s.End())
		for _, clause := range s.Body.List {
			caseClause := clause.(*ast.CaseClause)
			executionFromStmtList(builder, scopes.PushScope(), fset, caseClause.Body, returnOutputs).ConnectTo(afterSwitch.CurrentPoint)
		}
		return afterSwitch
	case *ast.TypeSwitchStmt:
		scopes = scopes.PushScope()
		builder = executionFromStmt(builder, scopes, fset, s.Init, returnOutputs)
		if assign, ok := s.Assign.(*ast.AssignStmt); ok && assign.Tok == token.DEFINE {
			utils.Assertf(len(assign.Lhs) == 1, "type switch assignment must have single variable")
			scopes.CreateVar(assign.Lhs[0].(*ast.Ident).Name)
		}
		afterSwitch := builder.AcquirePointWithRef(s.End())
		for _, clause := range s.Body.List {
			caseClause := clause.(*ast.CaseClause)
			executionFromStmtList(builder, scopes.PushScope(), fset, caseClause.Body, returnOutputs).ConnectTo(afterSwitch.CurrentPoint)
		}
		return afterSwitch
	case *ast.ReturnStmt:
		utils.Assertf(len(s.Results) == 0 || len(s.Results) == 1 || len(s.Results) == returnOutputs, "return inputs/outputs count mismatch: %v", fset.Position(s.Pos()))
		// naked return case
		if len(s.Results) == 0 {
			return builder.ApplyNextWithRef(ReturnVarsOp{VarIds: nil}, s.Pos())
		}
		resultOutputs := 1
		if len(s.Results) > 1 && returnOutputs == 1 {
			resultOutputs = returnOutputs
		}
		var varCompositions []VarComposition
		for _, result := range s.Results {
			var resultVarComposition []VarComposition
			builder, resultVarComposition = executionFromExpr(builder, scopes, fset, result, resultOutputs)
			varCompositions = append(varCompositions, resultVarComposition...)
		}
		varIds := make([]VarId, 0, len(s.Results))
		for i := range varCompositions {
			varId := scopes.NewVarId()
			varIds = append(varIds, varId)
			builder = executionFromVarComposition(fset, s, builder, VarSelector{VarId: varId}, varCompositions[i])
		}
		return builder.ApplyNextWithRef(ReturnVarsOp{VarIds: varIds}, s.Pos())
	case *ast.ExprStmt:
		builder, _ = executionFromExpr(builder, scopes, fset, s.X, 0)
		return builder
	case
		nil,
		*ast.BranchStmt, /* break / continue / goto */
		*ast.DeferStmt,
		*ast.EmptyStmt,
		*ast.GoStmt,
		*ast.IncDecStmt,
		*ast.SelectStmt,
		*ast.SendStmt:
		return builder
	}
	panic(fmt.Errorf("unexpected statement found: %T (%+v)", stmt, stmt))
}

func ExecutionFromFunc(
	scopes Scopes,
	fset *token.FileSet,
	funcDecl *ast.FuncDecl,
) Execution {
	builder := NewExecutionBuilder(fset)
	builder.AssignRef(builder.CurrentPoint, funcDecl.Pos())

	scopes = scopes.PushScope()
	if funcDecl.Type.Params != nil {
		for _, params := range funcDecl.Type.Params.List {
			for _, name := range params.Names {
				scopes.CreateVar(name.Name)
			}
		}
	}
	if funcDecl.Type.Results != nil {
		for _, params := range funcDecl.Type.Results.List {
			for _, name := range params.Names {
				scopes.CreateVar(name.Name)
			}
		}
	}
	builder = executionFromStmt(builder, scopes, fset, funcDecl.Body, funcDecl.Type.Results.NumFields())
	return builder.Build()
}

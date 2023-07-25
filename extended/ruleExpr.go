// Copyright 2021 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package extended

import (
	"github.com/antonmedv/expr"
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/types"
	"sync"
)

func ExprCompiledCall(code string, env map[string]interface{}) (interface{}, error) {
	program, err := expr.Compile(code, expr.Env(env))
	if err != nil {
		return nil, err
	}

	output, err := expr.Run(program, env)
	if err != nil {
		return nil, err
	}

	return output, nil
}

func ExprNonCompiledEval(code string, env map[string]interface{}) (interface{}, error) {
	output, err := expr.Eval(code, env)
	if err != nil {
		return nil, err
	}
	return output, nil
}

var mutexExpr sync.Mutex
var ExprContext = make(map[string]interface{})

func RegisterExprCustomFunc(name string, fn interface{}) {
	mutexExpr.Lock()
	defer mutexExpr.Unlock()
	ExprContext[name] = fn
}

func RegisterExprCustomFuncs(fns map[string]interface{}) {
	mutexExpr.Lock()
	defer mutexExpr.Unlock()

	for k, v := range fns {
		ExprContext[k] = v
	}
}

func DeleteExprCustomFunc(name string) {
	mutexExpr.Lock()
	defer mutexExpr.Unlock()
	delete(ExprContext, name)
}

func DeleteExprCustomFuncs(names []string) {
	mutexExpr.Lock()
	defer mutexExpr.Unlock()
	for _, name := range names {
		delete(ExprContext, name)
	}
}

func init() {
	rego.RegisterBuiltin2(
		&rego.Function{
			Name:             "Expr.NonCompiledEval",
			Decl:             types.NewFunction(types.Args(types.S, types.A), types.A),
			Memoize:          true,
			Nondeterministic: true,
		},
		func(bctx rego.BuiltinContext, code, env *ast.Term) (*ast.Term, error) {
			var exprCode string
			var input = make(map[string]interface{})

			if err := ast.As(code.Value, &exprCode); err != nil {
				return nil, err
			} else if err := ast.As(env.Value, &input); err != nil {
				return nil, err
			}

			ctx := make(map[string]interface{})
			for key, val := range ExprContext {
				ctx[key] = val
			}
			ctx["input"] = input

			output, err := ExprNonCompiledEval(exprCode, ctx)

			if err != nil {
				return nil, err
			}

			v, err := ast.InterfaceToValue(output)

			return ast.NewTerm(v), err
		},
	)

	rego.RegisterBuiltin2(
		&rego.Function{
			Name:             "Expr.CompiledEval",
			Decl:             types.NewFunction(types.Args(types.S, types.A), types.A),
			Memoize:          true,
			Nondeterministic: true,
		},
		func(bctx rego.BuiltinContext, code, env *ast.Term) (*ast.Term, error) {
			var exprCode string
			var input = make(map[string]interface{})

			if err := ast.As(code.Value, &exprCode); err != nil {
				return nil, err
			} else if err := ast.As(env.Value, &input); err != nil {
				return nil, err
			}

			ctx := make(map[string]interface{})
			for key, val := range ExprContext {
				ctx[key] = val
			}
			ctx["input"] = input

			output, err := ExprCompiledCall(exprCode, ctx)
			if err != nil {
				return nil, err
			}

			v, err := ast.InterfaceToValue(output)

			return ast.NewTerm(v), err
		},
	)
}

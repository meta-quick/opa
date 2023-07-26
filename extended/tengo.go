// Copyright 2021 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package extended

import (
	"context"
	"github.com/d5/tengo/v2"
	"github.com/meta-quick/opa/ast"
	"github.com/meta-quick/opa/rego"
	"github.com/meta-quick/opa/types"
	"sync"
)

func toEnviroments(vars map[string]interface{}) map[string]tengo.Object {
	if len(vars) == 0 {
		return nil
	}
	enviroment := make(map[string]tengo.Object)
	for k, v := range vars {
		enviroment[k] = toValue(v)
	}
	return enviroment
}

func TengoEval(script string, retkey string, env map[string]tengo.Object) (interface{}, error) {
	engine := tengo.NewScript([]byte(script))
	for k, val := range env {
		err := engine.Add(k, val)
		if err != nil {
			return nil, err
		}
	}

	compiled, err := engine.RunContext(context.Background())
	if err != nil {
		return nil, err
	}
	v := compiled.Get(retkey)

	return v.Value(), err
}

var mutex sync.Mutex
var TengoContext = make(map[string]interface{})

func RegisterTengoCustomFunc(name string, fn interface{}) {
	mutex.Lock()
	defer mutex.Unlock()
	TengoContext[name] = fn
}

func RegisterTengoCustomFuncs(fns map[string]interface{}) {
	mutex.Lock()
	defer mutex.Unlock()

	for k, v := range fns {
		TengoContext[k] = v
	}
}

func DeRegisterTengoCustomFunc(name string) {
	mutex.Lock()
	defer mutex.Unlock()
	delete(TengoContext, name)
}

func DeRegisterTengoCustomFuncs(names []string) {
	mutex.Lock()
	defer mutex.Unlock()
	for _, name := range names {
		delete(TengoContext, name)
	}
}

func init() {
	rego.RegisterBuiltin3(
		&rego.Function{
			Name:             "Tengo.Eval",
			Decl:             types.NewFunction(types.Args(types.S, types.S, types.A), types.A),
			Memoize:          true,
			Nondeterministic: true,
		},
		func(bctx rego.BuiltinContext, code, returned, env *ast.Term) (*ast.Term, error) {
			var exprCode string
			var retKey string
			var input = make(map[string]interface{})

			if err := ast.As(code.Value, &exprCode); err != nil {
				return nil, err
			} else if err := ast.As(returned.Value, &retKey); err != nil {
				return nil, err
			} else if err := ast.As(env.Value, &input); err != nil {
				return nil, err
			}

			ctx := toEnviroments(TengoContext)
			if ctx == nil {
				ctx = make(map[string]tengo.Object)
			}
			ctx["input"] = toValue(input)

			v, err := TengoEval(exprCode, retKey, ctx)
			if err != nil {
				return nil, err
			}
			vvv, err := ast.InterfaceToValue(v)
			return ast.NewTerm(vvv), err
		},
	)
}

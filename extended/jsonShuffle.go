package extended

import (
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/d5/tengo/v2"
	"github.com/meta-quick/opa/ast"
	"github.com/meta-quick/opa/rego"
	"github.com/meta-quick/opa/topdown"
	"github.com/meta-quick/opa/types"
)

func initTengoContext() {
	RegisterTengoCustomFunc("sprintf", fmt.Sprintf)
}

func copyTengoContext() map[string]tengo.Object {
	newContext := make(map[string]tengo.Object)
	for k, v := range TengoContext {
		newContext[k] = toValue(v)
	}
	return newContext
}

type aa struct {
	A    string
	Cost int
}

func JSONShuffle(model string, input *ast.Term) (*ast.Term, error) {
	//copy global context
	enviroment := copyTengoContext()

	var inputTengo = make(map[string]interface{})
	if err := ast.As(input.Value, &inputTengo); err != nil {
		return nil, err
	}

	target := input

	//init builtin variables
	aaa := aa{
		A:    "a",
		Cost: 100,
	}
	enviroment["input"] = toValue(inputTengo)
	enviroment["aa"] = toValue(aaa)

	//Unmarshal to map
	var shuffle = make(map[string]interface{}, 100)
	err := sonic.Unmarshal([]byte(model), &shuffle)
	if err != nil {
		return target, err
	}

	if shuffle != nil {
		// Remove denied fields
		for key, field := range shuffle {
			/* The input model is like this:
			 "filters": {
					"denied": [
						{
							"match": "a/b/c",
							"guard": "output := 1>1"
						}
					]
				},
			*/
			if key == "filters" {
				switch vv := field.(type) {
				case map[string]interface{}:
					for denied, column := range vv {
						if denied == "denied" {
							switch columns := column.(type) {
							case []interface{}:
								for _, exprs := range columns {
									switch expr := exprs.(type) {
									case map[string]interface{}:
										var match string
										var guard string
										for k, v := range expr {
											if k == "match" {
												match = v.(string)
											}
											if k == "guard" {
												guard = v.(string)
											}
										}

										retVal, err := TengoEval(guard, "output", enviroment)
										if err == nil {
											if retVal == true {
												//remove this field
												operations := make(map[string]string, 0)
												operations["op"] = "remove"
												operations["path"] = match

												patch, err := sonic.Marshal(operations)
												if err == nil {
													patches := ast.NewArray()
													patches = patches.Append(ast.MustParseTerm(string(patch)))
													target, err = topdown.ApplyPatches(target, patches)
													if err != nil {
														return target, err
													}
												}
											}
										}
									}
								}
							}
						}
					}
				}
			} //end if
		}
	}

	return target, nil
}

func init() {
	rego.RegisterBuiltin2(
		&rego.Function{
			Name:             "meta.json.shuffle",
			Decl:             types.NewFunction(types.Args(types.S, types.S), types.N),
			Memoize:          true,
			Nondeterministic: true,
		},
		func(bctx rego.BuiltinContext, model, target *ast.Term) (*ast.Term, error) {
			var template string
			if err := ast.As(model.Value, &template); err != nil {
				return nil, err
			}

			v, err := JSONShuffle(template, target)
			if err != nil {
				return nil, err
			}

			return v, err
		},
	)
}

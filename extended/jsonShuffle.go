package extended

import (
	"encoding/json"
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/d5/tengo/v2"
	"github.com/meta-quick/opa/ast"
	"github.com/meta-quick/opa/rego"
	"github.com/meta-quick/opa/topdown"
	"github.com/meta-quick/opa/topdown/builtins"
	"github.com/meta-quick/opa/types"
	"strconv"
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

type AA struct {
	A    string
	Cost int
}

func typeCasting(d interface{}) string {
	switch c := d.(type) {
	case string:
		return c
	case int32:
		return fmt.Sprint(c)
	case int64:
		return fmt.Sprint(c)
	case float64:
		return strconv.FormatFloat(c, 'f', -1, 64)
	case json.Number:
		tt, _ := strconv.ParseFloat(c.String(), 64)
		return strconv.FormatFloat(tt, 'f', -1, 64)
	default:
		return fmt.Sprint(d)
	}
}

func patch(op string) int {
	fmt.Println(op)
	return 0
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
	enviroment["input"] = toValue(toEnviroments(inputTengo))
	//below commented code is for testing
	//aa := &AA{
	//	A:    "a",
	//	Cost: 10,
	//}
	//enviroment["aa"] = toValue(aa)

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
			   "filters":{
			      "denied":[
			         {
			            "match":"code",
			            "guard":"input.x >=1"
			         }
			      ],
			      "rowfilter":{
			         "expr":"output := true; patch(\"hello world\")"
			      }
			   },
				"shuffle":{
					  "d/e/f":{
						 "algo":{
							"name":"mx.pfe.mask_string",
							"params":[
							   "2"
							],
							"guard":"1>1",
							"order":1
						 }
					  }
				   }
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
						if denied == "rowfilter" {
							switch columns := column.(type) {
							case map[string]interface{}:
								for k, v := range columns {
									if k == "expr" {
										_, err := TengoEval(v.(string), "output", enviroment)
										if err != nil {
											return target, err
										}
									}
								}
							}
						}
					} //end for
				}
			} //end if
		}

		//handle data shuffle such as mask
		for key, field := range shuffle {
			if key == "shuffle" {
				switch vv := field.(type) {
				case map[string]interface{}:
					for path, _ := range vv {
						fmt.Println(path)
						path := `aaa/:`
						pathV1, _ := ast.InterfaceToValue(path)
						regoPath, _ := topdown.ParsePath(ast.NewTerm(pathV1))
						Oobject, err := builtins.ObjectOperand(target.Value, 1)
						OOOO := Oobject.Get(ast.NewTerm(pathV1))

						println(OOOO)
						println(err)
						fmt.Println(regoPath)

						//switch algo := algoval.(type) {
						//case map[string]interface{}:
						//var name string
						//var params []interface{}
						//var guard string
						//for k, v := range algo {
						//	if k == "name" {
						//		name = v.(string)
						//	}
						//	if k == "params" {
						//		params = v.([]interface{})
						//	}
						//	if k == "guard" {
						//		guard = v.(string)
						//	}
						//}

						//retVal, err := TengoEval(guard, "output", enviroment)
						//if err == nil {
						//	if retVal == true {
						//		//remove this field
						//		operations := make(map[string]interface{}, 0)
						//		operations["op"] = "replace"
						//		operations["path"] = path
						//		operations["value"] = typeCasting(params[0])
						//
						//		patch, err := sonic.Marshal(operations)
						//		if err == nil {
						//			patches := ast.NewArray()
						//			patches = patches.Append(ast.MustParseTerm(string(patch)))
						//			target, err = topdown.ApplyPatches(target, patches)
						//			if err != nil {
						//				return target, err
						//			}
						//		}
						//	}
						//}
						//}
					}
				}
			}
		}
	}

	return target, nil
}

func init() {
	RegisterTengoCustomFunc("patch", patch)
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

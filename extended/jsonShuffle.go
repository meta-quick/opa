package extended

import (
	"encoding/json"
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/d5/tengo/v2"
	"github.com/meta-quick/mask/jsonmask"
	maskTypes "github.com/meta-quick/mask/types"
	"github.com/meta-quick/opa/ast"
	"github.com/meta-quick/opa/internal/edittree"
	"github.com/meta-quick/opa/rego"
	"github.com/meta-quick/opa/types"
	"github.com/spf13/cast"
	"strconv"
	"strings"
)

func initTengoContext() {
	RegisterTengoCustomFunc("sprintf", fmt.Sprintf)
	RegisterTengoCustomFunc("strContains", strings.Contains)
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

func expandPath(targetObj ast.Object, path *ast.Term) ([]ast.Ref, error) {
	// paths can either be a `/` separated json path or
	// an array or set of values
	var pathSegments ast.Ref
	switch p := path.Value.(type) {
	case ast.String:
		if p == "" {
			return nil, fmt.Errorf("empty path type ")
		}
		parts := strings.Split(strings.TrimLeft(string(p), "/"), "/")
		for _, part := range parts {
			part = strings.ReplaceAll(strings.ReplaceAll(part, "~1", "/"), "~0", "~")
			pathSegments = append(pathSegments, ast.StringTerm(part))
		}
	case *ast.Array:
		p.Foreach(func(term *ast.Term) {
			pathSegments = append(pathSegments, term)
		})
	default:
		return nil, fmt.Errorf("invalid path type %T", path.Value)
	}

	//expand path for array
	expandedSegments := make([]ast.Ref, 0)
	expandedSegmentsX := make([]ast.Ref, 0)

	for _, pathSegment := range pathSegments {
		switch pathSegment.Value.(type) {
		case ast.String:
			if pathSegment.Value.(ast.String) == ":" {
				//expand path
				//cache all expanded path
				expandedSegments = expandedSegmentsX
				//recreate expandedSegmentsX
				expandedSegmentsX = make([]ast.Ref, 0)
				for _, expandedSegment := range expandedSegments {
					parentObj, err := targetObj.Find(expandedSegment)
					if err == nil {
						switch parentObj.(type) {
						case *ast.Array:
							parentArray := parentObj.(*ast.Array)
							for i := 0; i < parentArray.Len(); i++ {
								extended := make(ast.Ref, 0)
								extended = append(extended, expandedSegment...)
								extended = append(extended, ast.IntNumberTerm(i))
								expandedSegmentsX = append(expandedSegmentsX, extended)
							}
						}
					}
				}

				//fixed expandedSegments is empty for the first time
				if len(expandedSegments) == 0 {
					for i := 0; i < targetObj.Len(); i++ {
						extended := make(ast.Ref, 0)
						extended = append(extended, ast.IntNumberTerm(i))
						expandedSegments = append(expandedSegments, extended)
					}
				}
			} else {
				//cache all expanded path
				expandedSegments = expandedSegmentsX
				//recreate expandedSegmentsX
				expandedSegmentsX = make([]ast.Ref, 0)
				for _, expandedSegment := range expandedSegments {
					extended := make(ast.Ref, 0)
					extended = append(extended, expandedSegment...)
					extended = append(extended, pathSegment)
					expandedSegmentsX = append(expandedSegmentsX, extended)
				}

				//fixed expandedSegments is empty for the first time
				if len(expandedSegments) == 0 {
					extended := make(ast.Ref, 0)
					extended = append(extended, pathSegment)
					expandedSegmentsX = append(expandedSegmentsX, extended)
				}
			}
		}
	}

	return expandedSegmentsX, nil
}

func PatchObject(keypatch map[string]string, target *ast.Term, et *edittree.EditTree, path string, rule interface{}, tengoEnviroment map[string]tengo.Object) {
	pathVal, err := ast.InterfaceToValue(path)
	if err != nil {
		return
	}

	targetObj, ok := target.Value.(ast.Object)
	if !ok {
		return
	}

	if rule == nil {
		return
	}

	guard := ""
	cbname := ""
	var args []string

	switch rule.(type) {
	case map[string]interface{}:
		ruleMap := rule.(map[string]interface{})
		for _, algo := range ruleMap {
			switch algo.(type) {
			case map[string]interface{}:
				xx := algo.(map[string]interface{})
				for k, v := range xx {
					switch v.(type) {
					case string:
						if k == "guard" {
							guard = v.(string)
						}

						if k == "name" {
							cbname = v.(string)
						}
					case int64:
					case float64:
					case []interface{}:
						if k == "params" {
							args, _ = cast.ToStringSliceE(v)
						}
					default:
						fmt.Println(k, v, "default")
					}
				}
			}
		}
	}

	if guard == "" {
		return
	}

	//Adjust args if cbname is for sm2 or sm4
	fn, args, canExec := AdjustMaskEvalArgs(cbname, args, keypatch)
	if !canExec {
		return
	}

	cb := jsonmask.ProcessHandle{
		Fn:   fn,
		Args: args,
	}

	expanded, err := expandPath(targetObj, ast.NewTerm(pathVal))
	if err != nil {
		return
	}

	for _, expandedSegment := range expanded {
		//check guard, make sure it is true. if not, skip
		var step ast.Value
		if guard != "" {
			step, err = targetObj.Find(expandedSegment)
			if err != nil {
				continue
			}

			vstep, err := ast.ValueToInterfaceX(step)
			if err != nil {
				continue
			}
			tengoEnviroment["Data"] = toValue(vstep)
			tengoEnviroment["DataPath"] = toValue(buildRefPath(expandedSegment))

			retVal, err := TengoEval(guard, "output", tengoEnviroment)
			if err != nil {
				continue
			}
			if retVal != true {
				continue
			}
		}

		if step == nil {
			step, err = targetObj.Find(expandedSegment)
			if err != nil {
				continue
			}
		}

		current, _ := ast.ValueToInterfaceX(step)
		ctx := maskTypes.BuiltinContext{
			Fn:      cb.Fn,
			Args:    cb.Args,
			Current: typeCasting(current),
		}

		maskTypes.Eval(&ctx)
		newValue, err := ast.InterfaceToValue(ctx.Result)

		_, err = et.DeleteAtPath(expandedSegment)
		if err != nil {
			continue
		}

		_, err = et.InsertAtPath(expandedSegment, ast.NewTerm(newValue))
		if err != nil {
			continue
		}
	}
}

func JSONShuffle(ns string, model string, input *ast.Term) (*ast.Term, error) {
	patchKeys := make(map[string]string, 2)
	//initial SM2 key if configured
	sm2cert := StoreGet(ns, SM2key)
	if &sm2cert != nil && sm2cert != "" {
		patchKeys[SM2key] = sm2cert
	}
	//initial SM4 key if configured
	sm4cert := StoreGet(ns, SM4key)
	if &sm4cert != nil && sm4cert != "" {
		patchKeys[SM4key] = sm4cert
	}

	//copy global context
	enviroment := copyTengoContext()

	var inputTengo = make(map[string]interface{})
	if err := ast.As(input.Value, &inputTengo); err != nil {
		return nil, err
	}

	target := input
	et := edittree.NewEditTree(target)
	targetObj, _ := target.Value.(ast.Object)

	//init builtin variables
	enviroment["input"] = toValue(toEnviroments(inputTengo))

	//Unmarshal to map
	var shuffle = make(map[string]interface{}, 100)
	err := sonic.Unmarshal([]byte(model), &shuffle)
	if err != nil {
		return target, err
	}

	// All algo execution at condition return true
	if shuffle != nil {
		// Do row filter at first,
		for key, field := range shuffle {
			if key == "filters" {
				switch vv := field.(type) {
				case map[string]interface{}:
					for denied, column := range vv {
						if denied == "rowfilter" {
							switch columns := column.(type) {
							case map[string]interface{}:
								for path, expr := range columns {
									vpath, _ := ast.InterfaceToValue(path)
									expanded, err := expandPath(targetObj, ast.NewTerm(vpath))
									if err != nil {
										continue
									}

									//loop through expanded path
									for i := len(expanded) - 1; i >= 0; i-- {
										expandedSegment := expanded[i]
										step, err := targetObj.Find(expandedSegment)
										if err != nil {
											continue
										}
										vstep, err := ast.ValueToInterfaceX(step)
										enviroment["Data"] = toValue(vstep)
										enviroment["DataPath"] = toValue(buildRefPath(expandedSegment))
										retVal, err := TengoEval(expr.(string), "output", enviroment)
										if err != nil {
											continue
										}
										//check condition if false , remove it, otherwise keep it
										if retVal != true {
											_, err = et.DeleteAtPath(expandedSegment)
											if err != nil {
												continue
											}
										}
									}
								}
							}
						}
					} //end for
				}
			} //end if
		}

		//reduce data size
		inputFiltered := et.Render()
		target = inputFiltered
		et = edittree.NewEditTree(target)
		targetObj, _ = target.Value.(ast.Object)

		// then, remove denied fields
		for key, field := range shuffle {
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
												//expand path
												vpath, _ := ast.InterfaceToValue(match)
												//TODO: Here should use filtered results
												expanded, err := expandPath(targetObj, ast.NewTerm(vpath))
												//remove this field
												if err == nil {
													//loop through expanded path
													for _, expandedSegment := range expanded {
														_, err = et.DeleteAtPath(expandedSegment)
														if err != nil {
															continue
														}
													}
												}
											}
										}
									}
								}
							}
						}
					} //end for
				}
			} //end if
		}

		//CHECKED: No data row removed at this point

		inputFiltered = et.Render()
		target = inputFiltered
		et = edittree.NewEditTree(target)
		targetObj, _ = target.Value.(ast.Object)

		//at last, handle data shuffle such as mask, encrypt, etc.
		for key, field := range shuffle {
			if key == "shuffle" {
				switch vv := field.(type) {
				case map[string]interface{}:
					for path, rule := range vv {
						PatchObject(patchKeys, target, et, path, rule, enviroment)
					}
				}
			}
		}
	}

	return et.Render(), nil
}

func AdjustMaskEvalArgs(fn string, args []string, keypatch map[string]string) (string, []string, bool) {
	if fn == maskTypes.SM2_MASK_STR.Name {
		//for SM2 case
		if sm2val, ok := keypatch[SM2key]; ok {
			if &sm2val == nil || sm2val == "" {
				//no set found, need skip, invalid case
				return fn, args, false
			}

			args = []string{sm2val}
			return fn, args, true
		}
	}

	if fn == maskTypes.SM4_MASK_STR.Name {
		//for SM4 case
		if sm4val, ok := keypatch[SM4key]; ok {
			if &sm4val == nil || sm4val == "" {
				//no set found, need skip, invalid case
				return fn, args, false
			}

			args = []string{sm4val}
			return fn, args, true
		}
	}

	return fn, args, true
}

func buildRefPath(path ast.Ref) string {
	var segments []string
	for _, segment := range path {
		segments = append(segments, segment.String())
	}
	return strings.Join(segments, "/")
}

func init() {
	RegisterTengoCustomFunc("patch", patch)
	rego.RegisterBuiltin3(
		&rego.Function{
			Name:             "meta.json.shuffle",
			Decl:             types.NewFunction(types.Args(types.S, types.S, types.S), types.N),
			Memoize:          true,
			Nondeterministic: true,
		},
		func(bctx rego.BuiltinContext, ns, model, target *ast.Term) (*ast.Term, error) {
			var template string
			if err := ast.As(model.Value, &template); err != nil {
				return nil, err
			}

			var namespace string
			if err := ast.As(ns.Value, &namespace); err != nil {
				return nil, err
			}

			v, err := JSONShuffle(namespace, template, target)
			if err != nil {
				return nil, err
			}

			return v, err
		},
	)
}

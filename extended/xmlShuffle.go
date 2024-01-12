package extended

import (
	"errors"
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/d5/tengo/v2"
	"github.com/duke-git/lancet/v2/convertor"
	"github.com/duke-git/lancet/v2/validator"
	"github.com/lestrrat-go/libxml2/parser"
	xmlTypes "github.com/lestrrat-go/libxml2/types"
	"github.com/lestrrat-go/libxml2/xpath"
	"github.com/meta-quick/mask/jsonmask"
	maskTypes "github.com/meta-quick/mask/types"
	"github.com/meta-quick/opa/ast"
	"github.com/meta-quick/opa/internal/edittree"
	"github.com/meta-quick/opa/rego"
	"github.com/meta-quick/opa/types"
	"github.com/spf13/cast"
)

func XMLShuffle(ns string, model string, input *ast.Term) (*ast.Term, error) {
	defer func() {
		if err := recover(); err != nil {
			//silent
		}
	}()

	// 获取待处理数据信息
	var inputMap = make(map[string]interface{})
	if err := ast.As(input.Value, &inputMap); err != nil {
		return nil, err
	}
	// 获取处理配置信息
	var config = make(map[string]interface{}, 100)
	err := sonic.Unmarshal([]byte(model), &config)
	if err != nil {
		return nil, err
	}
	// 设置SM算法秘钥参数
	smkeys := getSmKey(ns)
	// 获取待处理XML对象
	var rootDocument xmlTypes.Document
	xmlRspString := inputMap["rsp"]
	switch xmlRspString.(type) {
	case string:
		// 转换XML文本对象
		xmlParser := parser.New()
		rootDocument, err = xmlParser.Parse([]byte(xmlRspString.(string)))
		defer rootDocument.Free()
		if err != nil {
			return nil, err
		}
	default:
		// 不能处理，直接返回
		return input, nil
	}

	// 根据脱敏配置config进行脱敏处理
	newXmlRspString, err := doShuffle(inputMap, config, smkeys, rootDocument)
	if err != nil {
		return nil, err
	}

	// 修改 rsp 值为 处理后的 XML 字符串
	et, term, err, done := updateRspContext(err, newXmlRspString, input)
	if done {
		return term, err
	}
	return et.Render(), nil
}

func updateRspContext(err error, newXmlRspString string, input *ast.Term) (*edittree.EditTree, *ast.Term, error, bool) {
	newValue, err := ast.InterfaceToValue(newXmlRspString)
	et := edittree.NewEditTree(input)
	targetObj, ok := input.Value.(ast.Object)
	if !ok {
		return nil, nil, err, true
	}
	pathVal, err := ast.InterfaceToValue("rsp")
	if err != nil {
		return nil, nil, err, true
	}
	expanded, err := expandPath(targetObj, ast.NewTerm(pathVal))
	if err != nil {
		return nil, nil, err, true
	}
	for _, expandedSegment := range expanded {
		_, err = et.DeleteAtPath(expandedSegment)
		if err != nil {
			continue
		}
		_, err = et.InsertAtPath(expandedSegment, ast.NewTerm(newValue))
		if err != nil {
			continue
		}
	}
	return et, nil, nil, false
}

func doShuffle(inputMap map[string]interface{}, confInfo map[string]interface{},
	smKeys map[string]string, document xmlTypes.Document) (string, error) {
	if confInfo == nil {
		return "", errors.New("未解析到脱敏配置")
	}

	// rowfilter 数据行过滤
	for key, field := range confInfo {
		if key == "filters" {
			switch vv := field.(type) {
			case map[string]interface{}:
				for denied, column := range vv {
					if denied == "rowfilter" {
						switch columns := column.(type) {
						case map[string]interface{}:
							for path, expr := range columns {

								// 获取待处理XML节点数组循环删除
								nodes, err := findNodesFromPath(document, path)
								if err != nil {
									continue
								}
								for _, node := range nodes {
									// 先进行表达式判断，若表达式不成立，则直接跳过
									if !checkNeedDo(node, expr.(string)) {
										continue
									}
									// 删除节点
									parentNode, _ := node.ParentNode()
									parentNode.SetNodeValue("")
								}

							}
						}
					}
				}
			}
		}
	}

	// denied 数据列过滤
	for key, field := range confInfo {
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

									// 获取待处理XML节点数组循环删除
									nodes, err := findNodesFromPath(document, match)
									if err != nil {
										continue
									}
									for _, node := range nodes {
										// 先进行表达式判断，若表达式不成立，则直接跳过
										if !checkNeedDo(node, guard) {
											continue
										}
										// 删除节点
										parentNode, _ := node.ParentNode()
										parentNode.SetNodeValue("")
									}

								}
							}
						}
					}
				}
			}
		}
	}

	// shuffle数据脱敏
	for key, field := range confInfo {
		if key == "shuffle" {
			switch vv := field.(type) {
			case map[string]interface{}:
				for path, rule := range vv {
					changeXmlFromRule(path, rule, document, smKeys)
				}
			}
		}
	}

	// 返回结果字符串
	return document.String(), nil
}

func changeXmlFromRule(path string, rule interface{}, document xmlTypes.Document, smKeys map[string]string) {
	defer func() {
		if err := recover(); err != nil {
			//silent skip
		}
	}()
	if rule == nil {
		return
	}

	// 获取配置参数
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

	// SM2 && SM4 算法构建
	fn, args, canExec := AdjustMaskEvalArgs(cbname, args, smKeys)
	if !canExec {
		return
	}
	cb := jsonmask.ProcessHandle{
		Fn:   fn,
		Args: args,
	}

	nodes, err := findNodesFromPath(document, path)
	if err != nil {
		return
	}

	for _, node := range nodes {
		// 先进行表达式判断，若表达式不成立，则直接跳过
		if !checkNeedDo(node, guard) {
			continue
		}

		oldKey, oldValue := getNodeValueFromPath(node, path)

		// 函数处理
		ctx := maskTypes.BuiltinContext{
			Fn:      cb.Fn,
			Args:    cb.Args,
			Current: typeCasting(oldValue),
		}
		maskTypes.Eval(&ctx)
		newValue := ctx.Result

		// 结果修改
		changeXml(node, oldKey, convertor.ToString(newValue))
	}

}

func getNodeValueFromPath(node xmlTypes.Node, path string) (string, string) {
	// TODO 从path解析,取 内容 还是 属性

	//TODO 节点值
	value := node.NodeValue()
	println(value)
	//TODO 节点属性
	element := node.(xmlTypes.Element)
	attributes, _ := element.Attributes()
	for _, attribute := range attributes {
		sa := attribute.Value()
		println(sa)
	}

	return "", ""
}

func changeXml(node xmlTypes.Node, attrKey string, newV string) {
	if validator.IsEmptyString(attrKey) /* 内容 */ {
		node.SetNodeValue(newV)
	} else /* 属性 */ {
		// 改变节点属性值
		element := node.(xmlTypes.Element)
		attributes, err := element.Attributes()
		if err != nil {
			return
		}
		for _, attribute := range attributes {
			if attrKey == attribute.NodeName() {
				attribute.SetNodeValue(newV)
			}
		}
	}
}

func getNodeTypeFromPath(path string) string {

	return "context"
}

func checkNeedDo(node xmlTypes.Node, expr string) bool {
	// TODO Expr 表达式提取与转换

	// TODO 封装Tengo执行参数
	enviroment := make(map[string]tengo.Object)
	enviroment["Data"] = toValue("vstep") //TODO vstep是interface{}数据节点，这里我需要转换

	// 执行表达式判断
	r, err := TengoEval(expr, "output", enviroment)
	if r == nil || err != nil || r == false {
		return false
	}
	return true
}

func findNodesFromPath(document xmlTypes.Document, path string) (xmlTypes.NodeList, error) {
	ctx, _ := xpath.NewContext(document)
	// TODO 这里需要进行PATH转换

	// 注册命名空间
	prefix := `m`
	nsuri := `http://www.xyz.org/quotation`
	if err := ctx.RegisterNS(prefix, nsuri); err != nil {
		return nil, err
	}
	// 查询节点
	xPathResult, err := ctx.Find(`//m:Quotation` /*TODO 这里替换为解析后的 XPath */)
	return xpath.NodeList(xPathResult, err), nil
}

func getSmKey(ns string) map[string]string {
	//initial SM2 key if configured
	patchKeys := make(map[string]string, 2)
	sm2cert := StoreGet(ns, SM2key)
	if &sm2cert != nil && sm2cert != "" {
		patchKeys[SM2key] = sm2cert
	}
	//initial SM4 key if configured
	sm4cert := StoreGet(ns, SM4key)
	if &sm4cert != nil && sm4cert != "" {
		patchKeys[SM4key] = sm4cert
	}
	return patchKeys
}

func init() {
	rego.RegisterBuiltin3(
		&rego.Function{
			Name:             "meta.xml.shuffle",
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

			v, err := XMLShuffle(namespace, template, target)
			if err != nil {
				return nil, err
			}

			return v, err
		},
	)
}

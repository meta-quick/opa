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
	"strings"
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
	tengoContext := copyTengoContext()
	tengoContext["input"] = toValue(toEnviroments(inputMap))

	ctx, _ := xpath.NewContext(document)
	spaces := getNameSpaces(confInfo)
	for k, v := range spaces {
		// 注册命名空间
		_ = ctx.RegisterNS(strings.TrimSpace(k), strings.TrimSpace(v))
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
								nodes, err := findNodesFromPath(path, ctx)
								if err != nil {
									continue
								}
								for _, node := range nodes {
									// 先进行表达式判断，若表达式不成立，则直接跳过
									if !checkNeedDo(node, expr.(string), confInfo, tengoContext) {
										continue
									}
									// 删除节点
									parentNode, _ := node.ParentNode()
									_ = parentNode.RemoveChild(node)
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
									nodes, err := findNodesFromPath(match, ctx)
									if err != nil {
										continue
									}
									for _, node := range nodes {
										// 先进行表达式判断，若表达式不成立，则直接跳过
										if !checkNeedDo(node, guard, confInfo, tengoContext) {
											continue
										}
										// 删除节点
										parentNode, _ := node.ParentNode()
										_ = parentNode.RemoveChild(node)
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
					changeXmlFromRule(path, rule, document, smKeys, confInfo, tengoContext, ctx)
				}
			}
		}
	}

	// 返回结果字符串
	return document.String(), nil
}

func changeXmlFromRule(path string, rule interface{}, document xmlTypes.Document, smKeys map[string]string,
	confInfo map[string]interface{}, tengoContext map[string]tengo.Object, ctx *xpath.Context) {
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

	nodes, err := findNodesFromPath(path, ctx)
	if err != nil {
		return
	}

	for _, node := range nodes {
		// 先进行表达式判断，若表达式不成立，则直接跳过
		if !checkNeedDo(node, guard, confInfo, tengoContext) {
			continue
		}

		oldKey, oldValue := getNodeValueFromPath(node, path)
		if validator.IsEmptyString(oldKey) && validator.IsEmptyString(oldValue) {
			continue
		}

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
	// 从path解析,取 内容 还是 属性
	split := strings.Split(path, "/@")
	if len(split) > 1 {
		// 节点属性
		name := node.NodeName()
		if strings.EqualFold(name, split[1]) {
			value := node.NodeValue()
			return name, value
		}
	} else {
		// 节点值
		value := node.NodeValue()
		return "", value
	}
	return "", ""
}

func changeXml(node xmlTypes.Node, attrKey string, newV string) {
	if validator.IsEmptyString(attrKey) /* 内容 */ {
		node.SetNodeValue(newV)
	} else /* 属性 */ {
		// 改变节点属性值
		if strings.EqualFold(attrKey, node.NodeName()) {
			node.SetNodeValue(newV)
		}
	}
}

func checkNeedDo(node xmlTypes.Node, expr string, confInfo map[string]interface{}, tengoContext map[string]tengo.Object) bool {
	// 封装Tengo执行参数
	tengoContext["node"] = toValue(node.String())
	tengoContext["spaces"] = toValue(confInfo)
	// 执行表达式判断
	r, err := TengoEval(expr, "output", tengoContext)
	if r == nil || err != nil || r == false {
		return false
	}
	return true
}

func findNodesFromPath(path string, ctx *xpath.Context) (xmlTypes.NodeList, error) {
	// 查询节点
	xPath := strings.ReplaceAll(path, "/:", "")
	xPathResult, err := ctx.Find(xPath)
	return xpath.NodeList(xPathResult, err), nil
}

func getNameSpaces(confInfo map[string]interface{}) map[string]string {
	r := make(map[string]string)

	for key, field := range confInfo {
		if key == "namespaces" {
			switch vv := field.(type) {
			case []interface{}:
				for _, nameSpace := range vv {
					switch vvv := nameSpace.(type) {
					case map[string]interface{}:
						for k, v := range vvv {
							r[k] = convertor.ToString(v)
						}
					}
				}
			}
		}
	}
	return r
}

/*
var NameSpacePattern = `\{(.*?)}`
func getNameSpacesAndRemove(path string) (map[string]string, string) {
	// 创建正则表达式对象并编译模式
	re := regexp.MustCompile(NameSpacePattern)
	matchs := re.FindAllString(path, -1)
	r := make(map[string]string)
	for _, match := range matchs {
		path = strings.ReplaceAll(path, match, "")
		match = strings.Replace(match, "{", "", 1)
		match = strings.Replace(match, "}", "", len(match)-1)
		kv := strings.Split(match, "=")
		k := kv[0]
		v := kv[1]
		r[k] = v
	}
	return r, path
}
*/

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

	RegisterTengoCustomFunc("xml", getNodeValueFromXPath)
	RegisterTengoCustomFunc("sprintf", fmt.Sprintf)
	RegisterTengoCustomFunc("strContains", strings.Contains)
	RegisterTengoCustomFunc("strEqualFold", strings.EqualFold)
	RegisterTengoCustomFunc("toInt", toInt)
	RegisterTengoCustomFunc("toFloat", toFloat)
	RegisterTengoCustomFunc("toBool", toBool)
}

func toInt(v string) int64 {
	r, _ := convertor.ToInt(v)
	return r
}

func toFloat(v string) float64 {
	r, _ := convertor.ToFloat(v)
	return r
}

func toBool(v string) bool {
	r, _ := convertor.ToBool(v)
	return r
}

func getNodeValueFromXPath(nodeS string, path string, configInfo map[string]interface{}) string {
	docS := `<InnerHeader `
	spaces := getNameSpaces(configInfo)
	for k, v := range spaces {
		// 拼接命名空间 xmlns:m = "http://www.xyz.org/quotation"
		docS = docS + ` xmlns:` + k + ` = "` + v + `"`
	}
	docS = docS + `>` + nodeS + `</InnerHeader>`

	// 转换XML文本对象
	xmlParser := parser.New()
	rootDoc, err := xmlParser.Parse([]byte(docS))
	defer rootDoc.Free()
	if err != nil {
		return ""
	}

	ctx, _ := xpath.NewContext(rootDoc)
	for k, v := range spaces {
		// 注册命名空间
		_ = ctx.RegisterNS(strings.TrimSpace(k), strings.TrimSpace(v))
	}

	nodes, _ := findNodesFromPath(path, ctx)
	for _, node := range nodes {
		_, v := getNodeValueFromPath(node, path)
		return v
	}
	return ""
}

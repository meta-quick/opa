package extended

import (
	"github.com/bytedance/sonic"
	"github.com/lestrrat-go/libxml2/parser"
	xmlTypes "github.com/lestrrat-go/libxml2/types"
	"github.com/meta-quick/opa/ast"
	"github.com/meta-quick/opa/internal/edittree"
	"github.com/meta-quick/opa/rego"
	"github.com/meta-quick/opa/types"
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

func doShuffle(inputMap map[string]interface{}, config map[string]interface{},
	smKeys map[string]string, document xmlTypes.Document) (string, error) {
	// TODO 待处理逻辑

	return "处理后的XML字符串", nil
}

func getSmKey(ns string) map[string]string {
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

package tgx

import (
	"github.com/d5/tengo/v2/stdlib"
	"github.com/d5/tengo/v2"
	"fmt"
	"os"
	// "reflect"
	"context"
)

func NewTengo() *TengoX {
	return &TengoX{}
}

func (tgw *TengoX) LoadFile(path string, vars map[string]interface{}) (err error) {
	script, e := os.ReadFile(path)
	if e != nil {
		err = e
		return
	}
	return tgw.loadScript(script, vars)
}

func (tgw *TengoX) LoadScript(script string, vars map[string]interface{}) (err error) {
	return tgw.loadScript([]byte(script), vars)
}

func (tgw *TengoX) loadScript(script []byte, vars map[string]interface{}) (err error) {
	tgw.script = tengo.NewScript(script)
	vs := convertEnv(vars)
	tgw.addVars(vs)
	mods := stdlib.AllModuleNames()
	tgw.script.SetImports(stdlib.GetModuleMap(mods...))
	tgw.compiled, err = tgw.script.Run()
	return
}

func (tgw *TengoX) Run(vars map[string]interface{}) (err error) {
	if tgw.compiled == nil {
		err = fmt.Errorf("please load script first")
		return
	}
	if len(vars) > 0 {
		tgw.setVars(vars)
	}
	return tgw.compiled.Run()
}

func (tgw *TengoX) setVars(vars map[string]interface{}) {
	recompileNeeded := false
	vs := convertEnv(vars)
	for k, v := range vs {
		if err := tgw.compiled.Set(k, v); err != nil {
			recompileNeeded = true
			break
		}
	}
	if recompileNeeded {
		tgw.addVars(vs)
		tgw.compiled, _ = tgw.script.Run()
	}
}

func (tgw *TengoX) GetGlobal(name string) (res interface{}, err error) {
	res, err = tgw.getVar(name)
	return
}

func (tgw *TengoX) EvalFile(path string, params map[string]interface{}) (res interface{}, err error) {
	script, e := os.ReadFile(path)
	if e != nil {
		err = e
		return
	}
	return tgw.Eval(string(script), params)
}

func (tgw *TengoX) Eval(script string, params map[string]interface{}) (res interface{}, err error) {
	return tengo.Eval(context.Background(), script, params)
}

/*
func (tgw *TengoX) CallFunc(funcName string, args ...interface{}) (res interface{}, err error) {
	v, e := tgw.getVar(funcName)
	if e != nil {
		err = e
		return
	}
	fmt.Printf("v: %#v\n", v)
	fn, ok := v.(*tengo.UserFunction)
	if !ok {
		err = fmt.Errorf("var %s is not with type function", funcName)
		return
	}

	r, e := tgw.callFunc(fn, args...)
	if e != nil {
		err = e
		return
	}
	res = tengo.ToInterface(r)
	return
}

// bind a var of golang func with a tengo function name, so calling tengo function
// is just calling the related golang func.
// @param funcVarPtr  in format `var funcVar func(....) ...; funcVarPtr = &funcVar`
func (tgw *TengoX) BindFunc(funcName string, funcVarPtr interface{}) (err error) {
	if funcVarPtr == nil {
		err = fmt.Errorf("funcVarPtr must be a non-nil poiter of func")
		return
	}
	t := reflect.TypeOf(funcVarPtr)
	if t.Kind() != reflect.Ptr || t.Elem().Kind() != reflect.Func {
		err = fmt.Errorf("funcVarPtr expected to be a pointer of func")
		return
	}

	v, e := tgw.getVar(funcName)
	if e != nil {
		err = e
		return
	}
	fmt.Printf("-- v: %#v\n", v)
	fn, ok := v.(*tengo.UserFunction)
	if !ok {
		err = fmt.Errorf("var %s is not with type user-function", funcName)
		return
	}
	tgw.bindFunc(fn, funcVarPtr)
	return
}

// make a golang func as a built-in tengo function, so the function can be called in tengo script.
func (tgw *TengoX) MakeBuiltinFunc(funcName string, funcVar interface{}) (err error) {
	goFunc, e := bindGoFunc(funcName, funcVar)
	if e != nil {
		err = e
		return
	}
	tgw.script.Add(funcName, goFunc)
	return
}

// make a golang pointer of sturct instance as a tengo module.
// @param structVarPtr  pointer of struct instance is recommended.
func (tgw *TengoX) SetModule(modName string, structVarPtr interface{}) (err error) {
	if structVarPtr == nil {
		err = fmt.Errorf("structVarPtr must ba non-nil pointer of struct")
		return
	}
	v := reflect.ValueOf(structVarPtr)
	if v.Kind() == reflect.Struct || (v.Kind() == reflect.Ptr && v.Elem().Kind() == reflect.Struct) {
		tgw.script.Add(modName, bindGoStruct(v))
		return nil
	}
	err = fmt.Errorf("structVarPtr must be struct or pointer of strcut")
	return
}*/

func (tgw *TengoX) addVars(vs map[string]tengo.Object) {
	for k, v := range vs {
		if err := tgw.script.Add(k, v); err != nil {
			fmt.Printf("-- %v\n", err)
		}
	}
}

func convertEnv(vars map[string]interface{}) (map[string]tengo.Object) {
	if len(vars) == 0 {
		return nil
	}
	res := make(map[string]tengo.Object)
	for k, v := range vars {
		res[k] = toValue(v)
	}
	return res
}

func (tgw *TengoX) getVar(name string) (v interface{}, err error) {
	if tgw.compiled == nil {
		err = fmt.Errorf("no var named %s found", name)
		return
	}
	r := tgw.compiled.Get(name)
	if r == nil {
		err = fmt.Errorf("no var named %s found", name)
		return
	}
	v = r.Value()
	return
}

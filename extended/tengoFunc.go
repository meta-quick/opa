// Copyright 2021 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package extended

import (
	"fmt"
	"github.com/d5/tengo/v2"
	"reflect"
	"runtime"
	"strings"
)

func bindFunc(fn *tengo.UserFunction, funcVarPtr interface{}) {
	dest := reflect.ValueOf(funcVarPtr).Elem()
	fnType := dest.Type()
	dest.Set(reflect.MakeFunc(fnType, wrapFunc(fn, fnType)))
}

func wrapFunc(fn *tengo.UserFunction, fnType reflect.Type) func(args []reflect.Value) (results []reflect.Value) {
	return func(args []reflect.Value) (results []reflect.Value) {
		// make tengo args
		var tgArgs []tengo.Object
		lastNumIn := fnType.NumIn() - 1
		variadic := fnType.IsVariadic()
		for i, arg := range args {
			if i < lastNumIn || !variadic {
				tgArg, _ := tengo.FromInterface(arg.Interface())
				tgArgs = append(tgArgs, tgArg)
				continue
			}

			if arg.IsZero() {
				break
			}
			varLen := arg.Len()
			for j := 0; j < varLen; j++ {
				tgArg, _ := tengo.FromInterface(arg.Index(j).Interface())
				tgArgs = append(tgArgs, tgArg)
			}
		}

		// call tengo function
		res, err := fn.Call(tgArgs...)

		// convert result to golang
		results = make([]reflect.Value, fnType.NumOut())
		if err == nil {
			if fnType.NumOut() > 0 {
				if res.TypeName() == "array" {
					mRes := res.(*tengo.Array)
					n := fnType.NumOut()
					for i := 0; i < n; i++ {
						v := reflect.New(fnType.Out(i)).Elem()
						rv, err := mRes.IndexGet(&tengo.Int{Value: int64(i)})
						if err != nil {
							break
						}
						if err = setValue(v, tengo.ToInterface(rv)); err == nil {
							results[i] = v
						}
					}
				} else {
					v := reflect.New(fnType.Out(0)).Elem()
					rv := tengo.ToInterface(res)
					if err = setValue(v, rv); err == nil {
						results[0] = v
					}
				}
			}
		}

		if err != nil {
			nOut := fnType.NumOut()
			if nOut > 0 && fnType.Out(nOut-1).Name() == "error" {
				results[nOut-1] = reflect.ValueOf(err).Convert(fnType.Out(nOut - 1))
			} else {
				panic(err)
			}
		}

		for i, v := range results {
			if !v.IsValid() {
				results[i] = reflect.Zero(fnType.Out(i))
			}
		}

		return
	}
}

func callFunc(fn *tengo.UserFunction, args ...interface{}) (res tengo.Object, err error) {
	tgArgs := make([]tengo.Object, len(args))
	for i, arg := range args {
		tgArgs[i], _ = tengo.FromInterface(arg)
	}

	return fn.Call(tgArgs...)
}

func bindGoFunc(name string, funcVar interface{}) (goFunc *tengo.BuiltinFunction, err error) {
	if funcVar == nil {
		err = fmt.Errorf("funcVar must be a non-nil value")
		return
	}
	t := reflect.TypeOf(funcVar)
	if t.Kind() != reflect.Func {
		err = fmt.Errorf("funcVar expected to be a func")
		return
	}

	if len(name) == 0 {
		n := runtime.FuncForPC(reflect.ValueOf(funcVar).Pointer()).Name()
		if pos := strings.LastIndex(n, "."); pos >= 0 {
			name = n[pos+1:]
		} else {
			name = n
		}

		if len(name) == 0 {
			name = "noname"
		}
	}

	goFunc = &tengo.BuiltinFunction{Name: name, Value: wrapGoFunc(reflect.ValueOf(funcVar), t)}
	return
}

func wrapGoFunc(fnVal reflect.Value, fnType reflect.Type) tengo.CallableFunc {
	return func(args ...tengo.Object) (ret tengo.Object, err error) {
		// check args number
		argsNum := len(args)
		variadic := fnType.IsVariadic()
		lastNumIn := fnType.NumIn() - 1
		if variadic {
			if argsNum < lastNumIn {
				err = fmt.Errorf("at least %d args to call func", lastNumIn)
				return
			}
		} else {
			if argsNum != fnType.NumIn() {
				err = fmt.Errorf("%d args expected to call func", argsNum)
				return
			}
		}

		// make golang func args
		goArgs := make([]reflect.Value, argsNum)
		var fnArgType reflect.Type
		for i := 0; i < argsNum; i++ {
			if i < lastNumIn || !variadic {
				fnArgType = fnType.In(i)
			} else {
				fnArgType = fnType.In(lastNumIn).Elem()
			}

			goArgs[i] = makeValue(fnArgType)
			setValue(goArgs[i], tengo.ToInterface(args[i]))
		}

		// call golang func
		res := fnVal.Call(goArgs)

		// convert result to starlark
		retc := len(res)
		if retc == 0 {
			ret = tengo.UndefinedValue
			return
		}
		lastRetType := fnType.Out(retc - 1)
		if lastRetType.Name() == "error" {
			e := res[retc-1].Interface()
			if e != nil {
				err = e.(error)
				return
			}
			retc -= 1
			if retc == 0 {
				ret = tengo.UndefinedValue
				return
			}
		}

		if retc == 1 {
			ret, err = tengo.FromInterface(res[0].Interface())
			return
		}
		retV := make([]tengo.Object, retc)
		for i := 0; i < retc; i++ {
			if retV[i], err = tengo.FromInterface(res[i].Interface()); err != nil {
				return
			}
		}
		ret = &tengo.Array{Value: retV}
		return
	}
}

func bindGoStruct(structVar reflect.Value) (goModule *tengo.Map) {
	var structE reflect.Value
	if structVar.Kind() == reflect.Ptr {
		structE = structVar.Elem()
	} else {
		structE = structVar
	}
	structT := structE.Type()

	if structE == structVar {
		// struct is unaddressable, so make a copy of struct to an Elem of struct-pointer.
		// NOTE: changes of the copied struct cannot effect the original one. it is recommended to use the pointer of struct.
		structVar = reflect.New(structT) // make a struct pointer
		structVar.Elem().Set(structE)    // copy the old struct
		structE = structVar.Elem()       // structE is the copied struct
	}

	goModule = &tengo.Map{
		Value: wrapGoStruct(structVar, structE, structT),
	}
	return
}

func wrapGoStruct(structVar, structE reflect.Value, structT reflect.Type) map[string]tengo.Object {
	r := make(map[string]tengo.Object)
	for i := 0; i < structT.NumField(); i++ {
		strField := structT.Field(i)
		name := strField.Name
		name = strings.ToLower(name[:1]) + name[1:]
		fv := structE.Field(i)
		r[name], _ = tengo.FromInterface(fv.Interface())
	}

	// receiver is the struct
	bindGoMethod(structE, structT, r)

	// receiver is the pointer of struct
	t := structVar.Type()
	bindGoMethod(structVar, t, r)
	return r
}

func bindGoMethod(structV reflect.Value, structT reflect.Type, r map[string]tengo.Object) {
	for i := 0; i < structV.NumMethod(); i += 1 {
		m := structT.Method(i)
		name := strings.ToLower(m.Name[:1]) + m.Name[1:]
		mV := structV.Method(i)
		mT := mV.Type()
		r[name] = &tengo.BuiltinFunction{Name: name, Value: wrapGoFunc(mV, mT)}
	}
}

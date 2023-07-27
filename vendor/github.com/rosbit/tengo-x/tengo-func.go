package tgx

import (
	"github.com/d5/tengo/v2"
	"reflect"
)

func (tgw *TengoX) bindFunc(fn *tengo.UserFunction, funcVarPtr interface{}) {
	dest := reflect.ValueOf(funcVarPtr).Elem()
	fnType := dest.Type()
	dest.Set(reflect.MakeFunc(fnType, tgw.wrapFunc(fn, fnType)))
}

func (tgw *TengoX) wrapFunc(fn *tengo.UserFunction, fnType reflect.Type) func(args []reflect.Value) (results []reflect.Value) {
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
			for j:=0; j<varLen; j++ {
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
					for i:=0; i<n; i++ {
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
				results[nOut-1] = reflect.ValueOf(err).Convert(fnType.Out(nOut-1))
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

func (tgw *TengoX) callFunc(fn *tengo.UserFunction, args ...interface{}) (res tengo.Object, err error) {
	tgArgs := make([]tengo.Object, len(args))
	for i, arg := range args {
		tgArgs[i], _ = tengo.FromInterface(arg)
	}

	return fn.Call(tgArgs...)
}

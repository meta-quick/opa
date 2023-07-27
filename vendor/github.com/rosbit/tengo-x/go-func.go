package tgx

import (
	"github.com/d5/tengo/v2"
	"reflect"
	"fmt"
	"runtime"
	"strings"
)

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

	goFunc = &tengo.BuiltinFunction{Name:name, Value:wrapGoFunc(reflect.ValueOf(funcVar), t)}
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
		for i:=0; i<argsNum; i++ {
			if i<lastNumIn || !variadic {
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
		lastRetType := fnType.Out(retc-1)
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
		for i:=0; i<retc; i++ {
			if retV[i], err = tengo.FromInterface(res[i].Interface()); err != nil {
				return
			}
		}
		ret = &tengo.Array{Value:retV}
		return
	}
}

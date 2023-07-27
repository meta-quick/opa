package tgx

import (
	"github.com/d5/tengo/v2"
	"reflect"
	"strings"
)

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
	for i:=0; i<structT.NumField(); i++ {
		strField := structT.Field(i)
		name := strField.Name
		name = strings.ToLower(name[:1]) + name[1:]
		fv := structE.Field(i)
		r[name], _ = tengo.FromInterface(fv.Interface())
	}

	// receiver is the struct
	bindGoMethod(structE, structT, r)

	// reciver is the pointer of struct
	t := structVar.Type()
	bindGoMethod(structVar, t, r)
	return r
}

func bindGoMethod(structV reflect.Value, structT reflect.Type, r map[string]tengo.Object) {
	for i := 0; i<structV.NumMethod(); i+=1 {
		m := structT.Method(i)
		name := strings.ToLower(m.Name[:1]) + m.Name[1:]
		mV := structV.Method(i)
		mT := mV.Type()
		r[name] = &tengo.BuiltinFunction{Name:name, Value:wrapGoFunc(mV, mT)}
	}
}

package tgx

import (
	"github.com/d5/tengo/v2"
	"reflect"
	"fmt"
	"time"
)

func toValue(v interface{}) tengo.Object {
	if v == nil {
		return tengo.UndefinedValue
	}

	switch vv := v.(type) {
	case rune: // int32
		return &tengo.Char{Value:vv}
	case byte: // uint8
		return &tengo.Char{Value: rune(vv)}
	case int,int8,int16,int64:
		return &tengo.Int{Value:reflect.ValueOf(v).Int()}
	case uint,uint16,uint32:
		return &tengo.Int{Value:int64(reflect.ValueOf(v).Uint())}
	case uint64:
		if vv & 0x8000000000000000 == 0 {
			return &tengo.Int{Value:int64(vv)}
		}
		return &tengo.Float{Value:float64(vv)}
	case float32,float64:
		return &tengo.Float{Value:reflect.ValueOf(v).Float()}
	case string:
		if len(vv) > tengo.MaxStringLen {
			return &tengo.String{Value:vv[:tengo.MaxStringLen]}
		} else {
			return &tengo.String{Value:vv}
		}
	case []byte:
		if len(vv) > tengo.MaxBytesLen {
			return &tengo.Bytes{Value: vv[:tengo.MaxBytesLen]}
		}
		return &tengo.Bytes{Value:vv}
	case bool:
		if vv {
			return tengo.TrueValue
		}
		return tengo.FalseValue
	case time.Time:
		return &tengo.Time{Value:vv}
	case tengo.Object:
		return vv
	case error:
		return &tengo.Error{Value: toValue(vv.Error())}
	default:
		v2 := reflect.ValueOf(v)
		switch v2.Kind() {
		case reflect.Slice, reflect.Array:
			vt := v2
			r := make([]tengo.Object, vt.Len())
			for i:=0; i<vt.Len(); i++ {
				r[i] = toValue(vt.Index(i).Interface())
			}
			return &tengo.Array{Value:r}
		case reflect.Map:
			vm := v2
			r := make(map[string]tengo.Object)
			iter := vm.MapRange()
			for iter.Next() {
				k, v1 := iter.Key(), iter.Value()
				if k.Kind() == reflect.String {
					switch {
					case v1.IsNil():
						r[k.String()] = tengo.UndefinedValue
						continue
					case v1.Kind() == reflect.Func:
						if f, err := bindGoFunc(k.String(), v1.Interface()); err == nil {
							r[k.String()] = f
						}
						continue
					case v1.Kind() == reflect.Struct:
						r[k.String()] = bindGoStruct(v1)
						continue
					case v1.Kind() == reflect.Ptr && v1.Elem().Kind() == reflect.Struct:
						r[k.String()] = bindGoStruct(v1)
						continue
					default:
						if sv, ok := v1.Interface().(tengo.Object); ok {
							r[k.String()] = sv
							continue
						}
					}
				}
				r[k.String()] = toValue(v1.Interface())
			}
			return &tengo.Map{Value:r}
		case reflect.Struct:
			return bindGoStruct(v2)
		case reflect.Ptr:
			e := v2.Elem()
			if e.Kind() == reflect.Struct {
				return bindGoStruct(v2)
			}
			return toValue(e.Interface())
		case reflect.Func:
			if f, err := bindGoFunc("", v); err == nil {
				return f
			}
			return tengo.UndefinedValue
		default:
			return tengo.UndefinedValue
		}
	}
}

func setValue(dest reflect.Value, val interface{}) error {
	v := reflect.ValueOf(val)
	vt := reflect.TypeOf(val)
	dt := dest.Type()
	if vt.AssignableTo(dt) {
		dest.Set(v)
		return nil
	}

	if vt.ConvertibleTo(dt) {
		dest.Set(v.Convert(dt))
		return nil
	}

	return fmt.Errorf("cannot convert %s to %s", vt, dt)
}

func makeValue(t reflect.Type) reflect.Value {
	switch t.Kind() {
	case reflect.Slice:
		if t.Elem().Kind() == reflect.Uint8 {
			return reflect.Indirect(reflect.New(reflect.TypeOf("")))
		}
		fallthrough
	case reflect.Bool,reflect.Int,reflect.Uint,
			reflect.Int8,reflect.Int16,reflect.Int32,reflect.Int64,
			reflect.Uint8,reflect.Uint16,reflect.Uint32,reflect.Uint64,
			reflect.Float32,reflect.Float64,reflect.String,
			reflect.Array, reflect.Map, reflect.Struct:
		return reflect.Indirect(reflect.New(t))
	case reflect.Ptr:
		e := makeValue(t.Elem())
		return e.Addr()
	default:
		panic("unsupport type")
	}
}

func makeSlice(el reflect.Type) reflect.Value {
	t := reflect.SliceOf(el)
	return reflect.Indirect(reflect.New(t))
}

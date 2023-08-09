package types

import (
	"fmt"
	"strconv"
	"time"
)

func Eval(ctx *BuiltinContext) {
	_args := ctx.Args
	ctx.Args = make([]string,len(ctx.Args)+1)

	for i := 0; i < len(_args); i++ {
		ctx.Args[i+1] = _args[i]
	}
	ctx.Args[0] = ctx.Current

	if fn, ok := builtinFunctions[ctx.Fn]; ok {
		if decl,ok1 := BuiltinMap[ctx.Fn];ok1 {
			Args :=  decl.Decl.Args()
			args := buildArgs(Args,ctx)
			ctx.Result = fn(ctx,args)
		}
	}
}

func BuildBody(v interface{}) string {
	switch body := v.(type){
	case string:
		return body
	case float64:
		output := fmt.Sprintf("%f",body)
		return output
	case float32:
		output := fmt.Sprintf("%f",body)
		return output
	case int64:
		output := fmt.Sprintf("%d",body)
		return output
	case int:
		output := fmt.Sprintf("%d",body)
		return output
	case bool:
		output := "false"
		if body {
		   output = "true"
		}
		return output
	}

	return ""
}

func buildArgs(Args []Type,ctx *BuiltinContext) []interface{} {
	output := make([]interface{},len(Args))

	for i,arg := range Args {
		switch ty := arg.(type) {
		case String:
			output[i] = ctx.Args[i]
		case Number:
			output[i],ctx.Err = strconv.ParseInt(ctx.Args[i], 10, 64)
		case Boolean:
			output[i] = ctx.Args[i] == "true"
		case Float64:
			output[i],ctx.Err = strconv.ParseFloat(ctx.Args[i], 64)
		case Null:
			output[i] = nil
		case Number32:
			output[i],ctx.Err = strconv.Atoi(ctx.Args[i])
		case Date:
			output[i],ctx.Err = time.Parse(time.RFC3339,ctx.Args[i])
		case Any:
			output[i] = arg.(Any)
		default:
			fmt.Printf("%d  %T\n",i,ty)
		}
	}

	return output
}
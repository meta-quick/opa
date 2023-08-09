package types

import (
	"context"
	"github.com/meta-quick/mask/anonymity"
	"time"
)

type (
	Builtin struct {
		Name       string          `json:"name"`
		Decl       *Function `json:"decl"`
		Infix      string          `json:"infix,omitempty"`
		Relation   bool            `json:"relation,omitempty"`
		deprecated bool
	}

	BuiltinContext struct {
		Context                context.Context
		Current                string  //current content to be handled
		Fn                     string       //current mask process function
		Args                   []string//current mask process function
		Result				   interface{}  //return result
		Err                    error        //error if any
	}
	BuiltinFunc func(bctx *BuiltinContext, operands []interface{}) interface{}
)

var Builtins []*Builtin
var BuiltinMap map[string]*Builtin
func RegisterBuiltin(b *Builtin) {
	Builtins = append(Builtins, b)
	BuiltinMap[b.Name] = b
	if len(b.Infix) > 0 {
		BuiltinMap[b.Infix] = b
	}
}

var builtinFunctions = map[string]BuiltinFunc{}

func RegisterFunction(key string,fn BuiltinFunc) {
	builtinFunctions[key] = fn
}

var PFE_MASK_NUM = &Builtin{
	Name: "mx.pfe.mask_number",
	Decl: NewFunction(
		Args(
			N,
			N32,
		),
		N,
	),
}

var PFE_MASK_STR = &Builtin{
	Name: "mx.pfe.mask_string",
	Decl: NewFunction(
		Args(
			S,
			N32,
		),
		S,
	),
}

var DefaultBuiltins = []*Builtin {
	PFE_MASK_NUM,
	PFE_MASK_STR,
	HIDE_MASK_STR,
	HIDE_MASK_BOOLEAN,
	HIDE_MASK_FLOAT32,
	HIDE_MASK_FLOAT64,
	HIDE_MASK_INT32,
	HIDE_MASK_INT64,
	HIDE_MASK_DATESTRING,
	HIDE_MASK_DATE_MSEC,
	HIDE_MASK_STRX,
	FLOOR_MASK_FLOAT64,
	FLOOR_MASK_TIMEINMSEC,
	FLOOR_MASK_TIMESTRING,
}

var DefaultHandlerBuiltins = map[string]BuiltinFunc {
	PFE_MASK_NUM.Name: PFE_MASK_NUM_HANDLE,
	PFE_MASK_STR.Name: PFE_MASK_STR_HANDLE,
	HIDE_MASK_STR.Name: HIDING_MASK_STR_HANDLE,
	HIDE_MASK_BOOLEAN.Name: HIDING_MASK_BOOLEAN_HANDLE,
	HIDE_MASK_FLOAT32.Name: HIDING_MASK_FLOAT32_HANDLE,
	HIDE_MASK_FLOAT64.Name: HIDING_MASK_FLOAT64_HANDLE,
	HIDE_MASK_INT32.Name: HIDING_MASK_INT32_HANDLE,
	HIDE_MASK_INT64.Name: HIDING_MASK_INT64_HANDLE,
	HIDE_MASK_DATESTRING.Name: HIDING_MASK_TIME_HANDLE,
	HIDE_MASK_DATE_MSEC.Name: HIDING_MASK_TIMEMSEC_HANDLE,
	HIDE_MASK_STRX.Name: HIDING_MASK_STR0_HANDLE,
	FLOOR_MASK_FLOAT64.Name: FLOOR_MASK_FLOAT64_HANDLE,
	FLOOR_MASK_TIMEINMSEC.Name: FLOOR_MASK_TIMEMSEC_HANDLE,
	FLOOR_MASK_TIMESTRING.Name: FLOOR_MASK_TIME_HANDLE,
}

func init() {
	BuiltinMap = map[string]*Builtin{}

	for _, b := range DefaultBuiltins {
		RegisterBuiltin(b)
	}

	for k, v := range DefaultHandlerBuiltins {
		RegisterFunction(k, v)
	}
}

func PFE_MASK_NUM_HANDLE(bctx *BuiltinContext,args []interface{}) interface{} {
	pfe := anonymity.NewPrefixPreserveMasker()
	output,_ :=	pfe.MaskInteger(args[0].(int64),args[1].(int))
	return output
}

func PFE_MASK_STR_HANDLE(bctx *BuiltinContext,args []interface{}) interface{} {
	pfe := anonymity.NewPrefixPreserveMasker()
	output,_ :=	pfe.MaskString(args[0].(string),args[1].(int))
	return output
}

var HIDE_MASK_STR = &Builtin{
	Name: "mx.hide.mask_string",
	Decl: NewFunction(
		Args(
			S,
			S,
		),
		S,
	),
}

func HIDING_MASK_STR_HANDLE(bctx *BuiltinContext,args []interface{}) interface{} {
	hiding := anonymity.NewHidingMasker()
	output,_ :=	hiding.MaskString(args[0].(string),args[1].(string))
	return output
}

var HIDE_MASK_BOOLEAN = &Builtin{
	Name: "mx.hide.mask_boolean",
	Decl: NewFunction(
		Args(
			B,
		),
		B,
	),
}
func HIDING_MASK_BOOLEAN_HANDLE(bctx *BuiltinContext,args []interface{}) interface{} {
	hiding := anonymity.NewHidingMasker()
	output,_ :=	hiding.MaskBool(args[0].(bool))
	return output
}

var HIDE_MASK_FLOAT32 = &Builtin{
	Name: "mx.hide.mask_float32",
	Decl: NewFunction(
		Args(
			F,
		),
		F,
	),
}
func HIDING_MASK_FLOAT32_HANDLE(bctx *BuiltinContext,args []interface{}) interface{} {
	hiding := anonymity.NewHidingMasker()
	output,_ :=	hiding.MaskFloat32(args[0].(float32))
	return output
}

var HIDE_MASK_FLOAT64 = &Builtin{
	Name: "mx.hide.mask_float64",
	Decl: NewFunction(
		Args(
			F,
		),
		F,
	),
}
func HIDING_MASK_FLOAT64_HANDLE(bctx *BuiltinContext,args []interface{}) interface{} {
	hiding := anonymity.NewHidingMasker()
	output,_ :=	hiding.MaskFloat64(args[0].(float64))
	return output
}

var HIDE_MASK_INT32 = &Builtin{
	Name: "mx.hide.mask_int32",
	Decl: NewFunction(
		Args(
			N32,
		),
		N32,
	),
}
func HIDING_MASK_INT32_HANDLE(bctx *BuiltinContext,args []interface{}) interface{} {
	hiding := anonymity.NewHidingMasker()
	output,_ :=	hiding.MaskInt(args[0].(int))
	return output
}

var HIDE_MASK_INT64 = &Builtin{
	Name: "mx.hide.mask_int64",
	Decl: NewFunction(
		Args(
			N,
		),
		N,
	),
}
func HIDING_MASK_INT64_HANDLE(bctx *BuiltinContext,args []interface{}) interface{} {
	hiding := anonymity.NewHidingMasker()
	output,_ :=	hiding.MaskInt64(args[0].(int64))
	return output
}

func HIDING_MASK_UINT32_HANDLE(bctx *BuiltinContext,args []interface{}) interface{} {
	hiding := anonymity.NewHidingMasker()
	output,_ :=	hiding.MaskUint(args[0].(uint))
	return output
}

func HIDING_MASK_UINT64_HANDLE(bctx *BuiltinContext,args []interface{}) interface{} {
	hiding := anonymity.NewHidingMasker()
	output,_ :=	hiding.MaskUint64(args[0].(uint64))
	return output
}

var HIDE_MASK_DATESTRING = &Builtin{
	Name: "mx.hide.mask_timestring",
	Decl: NewFunction(
		Args(
			S,
		),
		S,
	),
}
func HIDING_MASK_TIME_HANDLE(bctx *BuiltinContext,args []interface{}) interface{} {
	hiding := anonymity.NewHidingMasker()
	tt,_ := time.Parse("2006-01-02 15:04:05",args[0].(string))
	output,_ :=	hiding.MaskTime(&tt)
	return output.Format("2006-01-02 15:04:05")
}

var HIDE_MASK_DATE_MSEC = &Builtin{
	Name: "mx.hide.mask_timemesc",
	Decl: NewFunction(
		Args(
			N,
		),
		N,
	),
}
func HIDING_MASK_TIMEMSEC_HANDLE(bctx *BuiltinContext,args []interface{}) interface{} {
	hiding := anonymity.NewHidingMasker()
	tt := time.UnixMilli(args[0].(int64))
	output,_ :=	hiding.MaskTime(&tt)
	return output.UnixMilli()
}

var HIDE_MASK_STRX = &Builtin{
	Name: "mx.hide.mask_strx",
	Decl: NewFunction(
		Args(
			S,
			S,
			N32,
			N32,
		),
		S,
	),
}
func HIDING_MASK_STR0_HANDLE(bctx *BuiltinContext,args []interface{}) interface{} {
	hiding := anonymity.NewHidingMasker()
	output,_ :=	hiding.MaskString0(args[0].(string),args[1].(string),args[2].(int),args[3].(int))
	return output
}

var FLOOR_MASK_FLOAT32 = &Builtin{
	Name: "mx.floor.mask_float32",
	Decl: NewFunction(
		Args(
			S,
			F,
		),
		F,
	),
}
func FLOOR_MASK_FLOAT32_HANDLE(bctx *BuiltinContext,args []interface{}) interface{} {
	floor := anonymity.NewFloorMasker()
	output,_ :=	floor.MaskFloat32(args[0].(float32))
	return output
}

var FLOOR_MASK_FLOAT64 = &Builtin{
	Name: "mx.floor.mask_float64",
	Decl: NewFunction(
		Args(
			F,
		),
		N,
	),
}
func FLOOR_MASK_FLOAT64_HANDLE(bctx *BuiltinContext,args []interface{}) interface{} {
	floor := anonymity.NewFloorMasker()
	output,_ :=	floor.MaskFloat64(args[0].(float64))
	return output
}

var FLOOR_MASK_TIMESTRING = &Builtin{
	Name: "mx.floor.mask_timestring",
	Decl: NewFunction(
		Args(
			S,
			S,
		),
		S,
	),
}
func FLOOR_MASK_TIME_HANDLE(bctx *BuiltinContext,args []interface{}) interface{} {
	floor := anonymity.NewFloorMasker()
	tt,_ := time.Parse("2006-01-02 15:04:05",args[0].(string))
	output,_ :=	floor.MaskTime(tt,args[1].(string))
	return output.Format("2006-01-02 15:04:05")
}

var FLOOR_MASK_TIMEINMSEC = &Builtin{
	Name: "mx.floor.mask_time_msec",
	Decl: NewFunction(
		Args(
			N,
			S,
		),
		N,
	),
}
func FLOOR_MASK_TIMEMSEC_HANDLE(bctx *BuiltinContext,args []interface{}) interface{} {
	floor := anonymity.NewFloorMasker()
	tt := time.UnixMilli(args[0].(int64))
	output,_ :=	floor.MaskTime(tt,args[1].(string))
	return output.UnixMilli()
}
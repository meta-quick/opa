package types

import (
	"fmt"
	"strings"
)

// Type represents a type of a term in the language.
type Type interface {
	String() string
	typeMarker() string
}

const (
	typeNull     = "null"
	typeBoolean  = "boolean"
	typeNumber   = "number"
	typeNum32   = "int32"
	typeDate   = "date"
	typeFloat64   = "float64"
	typeString   = "string"
	typeAny      = "any"
	typeFunction = "function"
)

func Sprint(x Type) string {
	if x == nil {
		return "???"
	}
	return x.String()
}

func (Null) typeMarker() string     { return typeNull }
func (Boolean) typeMarker() string  { return typeBoolean }
func (Number) typeMarker() string   { return typeNumber }
func (String) typeMarker() string   { return typeString }
func (Float64) typeMarker() string   { return typeFloat64 }
func (Number32) typeMarker() string   { return typeNum32 }
func (Date) typeMarker() string   { return typeDate }
func (Any) typeMarker() string      { return typeAny }
func (Function) typeMarker() string { return typeFunction }

// Null represents the null type.
type Null struct{}

// NewNull returns a new Null type.
func NewNull() Null {
	return Null{}
}
func (t Null) String() string {
	return typeNull
}

// Boolean represents the boolean type.
type Boolean struct{}

// B represents an instance of the boolean type.
var B = NewBoolean()

// NewBoolean returns a new Boolean type.
func NewBoolean() Boolean {
	return Boolean{}
}
func (t Boolean) String() string {
	return t.typeMarker()
}

// String represents the string type.
type String struct{}

// S represents an instance of the string type.
var S = NewString()

// NewString returns a new String type.
func NewString() String {
	return String{}
}
func (t String) String() string {
	return typeString
}

// Number represents the number type.
type Number struct{}

// N represents an instance of the number type.
var N = NewNumber()

// NewNumber returns a new Number type.
func NewNumber() Number {
	return Number{}
}
func (Number) String() string {
	return typeNumber
}

// Float64 represents the Float64 type.
type Float64 struct{}

// F represents an instance of the Float64 type.
var F = NewFloat64()

// NewFloat64 returns a new Float64 type.
func NewFloat64() Float64 {
	return Float64{}
}
func (Float64) String() string {
	return typeFloat64
}

// Number32 represents the Number32 type.
type Number32 struct{}

// N32 represents an instance of the Number32 type.
var N32 = NewNumber32()

// NewNumber32 returns a new Number32 type.
func NewNumber32() Number32 {
	return Number32{}
}

func (Number32) String() string {
	return typeNum32
}

// Date represents the date type.
type Date struct{}

// D represents an instance of the date type.
var D = NewDate()

// NewDate returns a new date type.
func NewDate() Date {
	return Date{}
}
func (Date) String() string {
	return typeDate
}

// Any represents a dynamic type.
type Any []Type

// A represents the superset of all types.
var A = NewAny()

// NewAny returns a new Any type.
func NewAny(of ...Type) Any {
	sl := make(Any, len(of))
	for i := range sl {
		sl[i] = of[i]
	}
	return sl
}
func (t Any) String() string {
	prefix := "any"
	if len(t) == 0 {
		return prefix
	}
	buf := make([]string, len(t))
	for i := range t {
		buf[i] = Sprint(t[i])
	}
	return prefix + "<" + strings.Join(buf, ", ") + ">"
}

// Function represents a function type.
type Function struct {
	args     []Type
	result   Type
	variadic Type
}

// Args returns an argument list.
func Args(x ...Type) []Type {
	return x
}

// NewFunction returns a new Function object where xs[:len(xs)-1] are arguments
func NewFunction(args []Type, result Type) *Function {
	return &Function{
		args:   args,
		result: result,
	}
}

// NewVariadicFunction returns a new Function object. This function sets the
// variadic bit on the signature. Non-void variadic functions are not currently
// supported.
func NewVariadicFunction(args []Type, varargs Type, result Type) *Function {
	if result != nil {
		panic("illegal value: non-void variadic functions not supported")
	}
	return &Function{
		args:     args,
		variadic: varargs,
		result:   nil,
	}
}

// FuncArgs returns the function's arguments.
func (t *Function) FuncArgs() FuncArgs {
	return FuncArgs{Args: t.Args(), Variadic: t.variadic}
}

// Args returns the function's arguments as a slice, ignoring variadic arguments.
func (t *Function) Args() []Type {
	cpy := make([]Type, len(t.args))
	copy(cpy, t.args)
	return cpy
}

// Result returns the function's result type.
func (t *Function) Result() Type {
	return t.result
}
func (t *Function) String() string {
	return fmt.Sprintf("%v => %v", t.FuncArgs(), Sprint(t.Result()))
}

// FuncArgs represents the arguments that can be passed to a function.
type FuncArgs struct {
	Args     []Type `json:"args,omitempty"`
	Variadic Type   `json:"variadic,omitempty"`
}

func (a FuncArgs) Arg(x int) Type {
	if x < len(a.Args) {
		return a.Args[x]
	}
	return a.Variadic
}
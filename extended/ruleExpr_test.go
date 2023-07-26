// Copyright 2021 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package extended

import (
	"context"
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/meta-quick/opa/rego"
	"testing"
)

func TestExpr(t *testing.T) {
	env := map[string]interface{}{
		"greet":   "Hello, %v!",
		"names":   []string{"world", "you"},
		"sprintf": fmt.Sprintf,
	}

	code := `sprintf(greet, names[0])`

	output, _ := ExprCompiledCall(code, env)

	fmt.Println(output)
}

func TestExprNoCompilied(t *testing.T) {
	env := map[string]interface{}{
		"greet":   "Hello, %v!",
		"names":   []string{"world", "you"},
		"sprintf": fmt.Sprintf,
	}

	code := `sprintf(greet, names[0])`

	output, _ := ExprNonCompiledEval(code, env)

	fmt.Println(output)
}

func TestExprRego(t *testing.T) {
	code := `sprintf("hello %v",greet)`

	input := map[string]interface{}{
		"code": code,
	}

	ctx := context.Background()

	module := "\t\tpackage test\n\t\tp = output {\n        xxx := Expr.CompiledEval(`%s`,input)\n        output := input\n}"
	module = fmt.Sprintf(module, code)

	r := rego.New(
		rego.Query("data.test.p"),
		rego.Module("", module),
	)

	RegisterExprCustomFunc("sprintf", fmt.Sprintf)
	RegisterExprCustomFunc("greet", "demo")

	pq, err := r.PrepareForEval(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err.Error())
	}

	if err != nil {
		t.Fatal(err)
	}

	//xxx, _ := sonic.Marshal(input)

	output, err := pq.Eval(ctx, rego.EvalInput(input))
	ret, _ := sonic.Marshal(output)
	fmt.Println(string(ret[:]))
}

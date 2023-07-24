// Copyright 2021 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package extended

import (
	"context"
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/open-policy-agent/opa/rego"
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
	code := `"hello " + greet`
	env := map[string]interface{}{
		"greet":   "world",
		"sprintf": "fmt.Sprintf",
	}

	input := map[string]interface{}{
		"code":    code,
		"context": env,
	}

	ctx := context.Background()

	module := `
		package test
		p = output {
        #x := timed.Gauge.Add("ns","ley",100,100000)
        xxx := Expr.CompiledEval(input.code,input.context)
        output := input
}
`
	r := rego.New(
		rego.Query("data.test.p"),
		rego.Module("", module),
	)

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

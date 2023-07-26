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

func adder(a, b int) int {
	return a + b
}

func TestRegoTengo(t *testing.T) {
	code := "each := func(seq, fn) {\n    for x in seq { fn(x) }\n}\n\nsum := 0\nmul := 1\neach([1, 2, 3, 4], func(x) {\n    sum += x\n    mul *= x\n})\ninput.x = 200\noutput := adder(100,2000)"
	env := map[string]interface{}{
		"greet":   "world",
		"sprintf": "fmt.Sprintf",
	}

	input := map[string]interface{}{
		"code":    "code",
		"context": env,
		"x":       100,
	}

	RegisterTengoCustomFunc("adder", adder)

	// set values
	module := "\t\tpackage test\n\t\tp = output {\n        xxxx := Tengo.Eval(`%s`,\"output\",input.context)\n        output := xxxx\n}"
	module = fmt.Sprintf(module, code)

	ctx := context.Background()

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

package extended

import (
	"context"
	"github.com/bytedance/sonic"
	"github.com/meta-quick/opa/ast"
	"github.com/meta-quick/opa/rego"
	"testing"
)

func TestPersist(t *testing.T) {
	var key, _ = ast.InterfaceToValue("tk")
	var val, _ = ast.InterfaceToValue(100)
	var ns, _ = ast.InterfaceToValue("demo")
	var dr, _ = ast.InterfaceToValue(1000)

	_, err := GaugeAdd(ns, key, val, dr)
	if err != nil {
		return
	}
	_, err = GaugeAdd(ns, key, val, dr)
	if err != nil {
		return
	}

	_, err = GaugeGet(ns, key)
	if err != nil {
		return
	}
}

func TestSonic(t *testing.T) {
	input := []byte(`{"key1":[{},{"key2":{"key3":[1,2,3]}}]}`)
	var vv = make(map[string]interface{}, 100)

	sonic.Unmarshal(input, &vv)

	xx, err := sonic.Marshal(vv)

	println(xx, err)
}

func TestPersistRego(t *testing.T) {
	code := `sprintf(greet, names[0])`
	env := map[string]interface{}{
		"greet":   "Hello, %v!",
		"names":   []string{"world", "you"},
		"sprintf": "fmt.Sprintf",
	}

	input := map[string]interface{}{
		"code": code,
		"env":  env,
	}

	xx, err := sonic.Marshal(input)

	if err == nil {
		println(xx)
	}

	//ShuffleModelAddString("oa", "api", model)
	ctx := context.Background()

	module := `
		package test
		p = output {
		x := timed.Gauge.Add("ns","ley",100,100000)
        xx := timed.Gauge.Add("ns","ley",100,100000)
        output := xx
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

	output, err := pq.Eval(ctx, rego.EvalInput(string(xx)))
	vvx, _ := sonic.Marshal(output)
	println(string(vvx))
}

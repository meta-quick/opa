package extended

import (
	"context"
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/meta-quick/opa/rego"
	"io/ioutil"
	"testing"
)

func TestJSONShuffle(t *testing.T) {
	model := `{
   "filters":{
      "denied":[
         {
            "match":"code",
            "guard":"output := input.x ==1"
         }
      ],
      "rowfilter":{
         "records/:":"output := true"
      }
   },
"shuffle":{
      "records/:/contact/:":{
         "algo":{
            "name":"mx.pfe.mask_string",
            "params":[
               "2"
            ],
            "guard":"output := 1>0",
            "order":1
         }
      }
   }
}`

	env := map[string]interface{}{
		"greet":   "world",
		"sprintf": "fmt.Sprintf",
	}

	contact1 := []interface{}{
		"zhangsan@qq.com",
		"123456789",
	}

	contact2 := []interface{}{
		"lishi@qq.com",
		"123456789",
	}

	r1 := map[string]interface{}{
		"nane":    "zhangsan",
		"age":     18,
		"contact": contact1,
	}

	r2 := map[string]interface{}{
		"name":    "lishi",
		"age":     18,
		"contact": contact2,
	}

	array := []interface{}{r1, r2}

	input := map[string]interface{}{
		"code":    0,
		"message": "success",
		"records": array,
		"context": env,
		"x":       100,
	}

	fmt.Println(model)
	// set values  xxxx := object.get(input,["xxx",0],0)
	//  xxxx := meta.json.shuffle(`%s`,input)
	module := "\t\tpackage test\n\t\tp = output {\n       xxxx := meta.json.shuffle(\"nscc\",`%s`,input)\n        output := xxxx\n}"
	module = fmt.Sprintf(module, model)

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

func Test_Main(t *testing.T) {
	rsp, _ := ioutil.ReadFile("data.json")

	module := "package play\nimport future.keywords.in\ndefault pre = false\ndefault main = {}\npre {\n    input.reqm.withPII == input.objm.withPII\n    some i, j\n    input.reqm.__builtin_g[i] == input.objm.__builtin_g[j]\n}\nmain() = rsp {\n    rsp := meta.json.shuffle(\"nscc\",\"{\\\"filters\\\":{\\\"rowfilter\\\":{\\\"rsp/value/:\\\":\\\"output := Data.age == 18\\\"},\\\"denied\\\":[{\\\"guard\\\":\\\"output := true \\\",\\\"match\\\":\\\"rsp/value/:/phone\\\"}]},\\\"shuffle\\\":{\\\"rsp/value/:/password\\\":{\\\"algo\\\":{\\\"guard\\\":\\\"output := true\\\",\\\"params\\\":[\\\"0\\\"],\\\"name\\\":\\\"mx.pfe.mask_string\\\"}}}}\", input)\n}"
	ctx := context.Background()

	r := rego.New(
		rego.Query("data.play.main"),
		rego.Module("", module),
	)

	pq, err := r.PrepareForEval(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err.Error())
	}

	if err != nil {
		t.Fatal(err)
	}

	var rspObj interface{}
	sonic.Unmarshal([]byte(rsp), &rspObj)

	input := map[string]interface{}{
		"rsp": rspObj,
	}
	output, err := pq.Eval(ctx, rego.EvalInput(input))
	ret, err := sonic.Marshal(output[0].Expressions[0].Value)

	fmt.Println(string(ret[:]))
}

package extended

import (
	"context"
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/meta-quick/opa/rego"
	"testing"
)

func TestJSONShuffle(t *testing.T) {
	model := `{
   "filters":{
      "denied":[
         {
            "match":"code",
            "guard":"input.x >=1"
         }
      ],
      "rowfilter":{
         "expr":"output := true; patch(\"hello world\")"
      }
   },
"shuffle":{
      "d/e/f":{
         "algo":{
            "name":"mx.pfe.mask_string",
            "params":[
               "2"
            ],
            "guard":"1>1",
            "order":1
         }
      }
   }
}`

	env := map[string]interface{}{
		"greet":   "world",
		"sprintf": "fmt.Sprintf",
	}

	array := []interface{}{"hello", "world"}

	input := map[string]interface{}{
		"code":    "helloworld",
		"xxx":     array,
		"context": env,
		"x":       100,
	}

	fmt.Println(model)
	// set values  xxxx := object.get(input,["xxx",0],0)
	//  xxxx := meta.json.shuffle(`%s`,input)
	module := "\t\tpackage test\n\t\tp = output {\n       xxxx := meta.json.shuffle(`%s`,input)\n        output := xxxx\n}"
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

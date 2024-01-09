package extended

import (
	"context"
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/meta-quick/opa/rego"
	"testing"
)

func TestXMLShuffle(t *testing.T) {
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

	input := map[string]interface{}{
		"code":    0,
		"message": "success",
		"x":       100,
		"rsp": `
<SOAP-ENV:Envelope
        xmlns:SOAP-ENV = "http://www.w3.org/2001/12/soap-envelope"
        SOAP-ENV:encodingStyle = "http://www.w3.org/2001/12/soap-encoding">

    <SOAP-ENV:Body xmlns:m = "http://www.xyz.org/quotation">
        <m:GetQuotationResponse>
            <m:Quotation Name="MY_TEST" >Here is the quotation</m:Quotation>
        </m:GetQuotationResponse>
    </SOAP-ENV:Body>
</SOAP-ENV:Envelope>
`,
	}

	fmt.Println(model)
	// set values  xxxx := object.get(input,["xxx",0],0)
	//  xxxx := meta.json.shuffle(`%s`,input)
	module := "\t\tpackage test\n\t\tp = output {\n       xxxx := meta.xml.shuffle(\"nscc\",`%s`,input)\n        output := xxxx\n}"
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

func Test_XML_Main(t *testing.T) {

}

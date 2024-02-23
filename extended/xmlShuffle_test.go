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
	"namespaces":[
		{"SOAP-ENV":"http://www.w3.org/2001/12/soap-envelope"},
		{"m":"http://www.xyz.org/quotation"}
    ],
   "filters":{
      "denied":[
         {
            "match":"//SOAP-ENV:Envelope/:/SOAP-ENV:Body/:/m:GetQuotationResponse/:/m:Code[@id=\"ww\"]",
            "guard":"na :=  xml(node,\"//m:Code/m:Name/@id\",spaces) \n output := strEqualFold(na,\"na\") "
         }
      ],
      "rowfilter":{
         "//SOAP-ENV:Envelope/:/SOAP-ENV:Body/:/m:GetQuotationResponse/:/m:Quotation[@id=\"ls\"]":"output := true"
      }
   },
"shuffle":{
      "//SOAP-ENV:Envelope/:/SOAP-ENV:Body/:/m:GetQuotationResponse/:/m:Quotation[@id=\"zs\"]/@name":{
         "algo":{
            "name":"mx.pfe.mask_string",
            "params":[
               "6"
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
            <m:Quotation id="zs" name="ZS_ATTR_TEXT_CONTEXT" >ZS_TEST_CONTEXT</m:Quotation>
            <m:Quotation id="ls" >LS</m:Quotation>
			<m:Code id="ww">
				<m:Name id="na">name</m:Name>
				<m:Age>18</m:Age>
			</m:Code>
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

func TestXMLShuffleAutoTransfer(t *testing.T) {
	model := `{
	"namespaces":[
		{"SOAP-ENV":"http://www.w3.org/2001/12/soap-envelope"},
		{"m":"http://www.xyz.org/quotation"}
    ],
   "filters":{
      "denied":[
         {
            "match":"//SOAP-ENV:Envelope/:/SOAP-ENV:Body/:/m:GetQuotationResponse/:/m:Code[@id=\"ww\"]",
            "guard":"na :=  xml(node,\"//m:Code/m:Name/@id\",spaces) \n output := strEqualFold(na,\"na\") "
         }
      ],
      "rowfilter":{
         "//SOAP-ENV:Envelope/:/SOAP-ENV:Body/:/m:GetQuotationResponse/:/m:Quotation[@id=\"ls\"]":"output := true"
      }
   },
"shuffle":{
      "//SOAP-ENV:Envelope/:/SOAP-ENV:Body/:/m:GetQuotationResponse/:/m:Quotation[@id=\"zs\"]":{
         "algo":{
            "name":"mx.hide.mask_strx",
            "params":[
               "<!--","1","3"
            ],
            "guard":"output := 1 > 0",
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
            <m:Quotation id="zs" name="ZS_ATTR_TEXT_CONTEXT" >ZS_TEST_CONTEXT</m:Quotation>
            <m:Quotation id="ls" >LS</m:Quotation>
			<m:Code id="ww">
				<m:Name id="na">name</m:Name>
				<m:Age>18</m:Age>
			</m:Code>
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

func TestXMLShuffleConvertor(t *testing.T) {
	model := `{
	"namespaces":[
		{"SOAP-ENV":"http://www.w3.org/2001/12/soap-envelope"},
		{"m":"http://www.xyz.org/quotation"}
    ],
   "filters":{
      "denied":[
         {
            "match":"//SOAP-ENV:Envelope/:/SOAP-ENV:Body/:/m:GetQuotationResponse/:/m:Code[@id=\"ww\"]",
            "guard":" output := ( toNumber(xml(node,\"//m:Code/m:Age\",spaces)) == 18.0 ) "
         }
      ],
      "rowfilter":{
         "//SOAP-ENV:Envelope/:/SOAP-ENV:Body/:/m:GetQuotationResponse/:/m:Quotation[@id=\"ls\"]":"output := true"
      }
   },
"shuffle":{
      "//SOAP-ENV:Envelope/:/SOAP-ENV:Body/:/m:GetQuotationResponse/:/m:Quotation[@id=\"zs\"]/@name":{
         "algo":{
            "name":"mx.pfe.mask_string",
            "params":[
               "6"
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
            <m:Quotation id="zs" name="ZS_ATTR_TEXT_CONTEXT" >ZS_TEST_CONTEXT</m:Quotation>
            <m:Quotation id="ls" >LS</m:Quotation>
			<m:Code id="ww">
				<m:Name id="na">name</m:Name>
				<m:Age>18.0</m:Age>
			</m:Code>
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

func TestXMLShuffle_2(t *testing.T) {
	model := `{
	"namespaces":[
		{"SOAP-ENV":"http://www.w3.org/2001/12/soap-envelope"},
		{"m":"http://www.xyz.org/quotation"}
    ],
   "filters":{
      "denied":[
         {
            "match":"//SOAP-ENV:Envelope/:/SOAP-ENV:Body/:/m:GetQuotationResponse/:/m:Code[@id=\"ww\"]",
            "guard":"na :=  xml(node,\"//m:Code/m:Name/@id\",spaces) \n output := strEqualFold(na,\"na\") "
         }
      ],
      "rowfilter":{
         "//SOAP-ENV:Envelope/:/SOAP-ENV:Body/:/m:GetQuotationResponse/:/m:Quotation[@id=\"ls\"]":"output := true"
      }
   },
"shuffle":{
      "//SOAP-ENV:Envelope/:/SOAP-ENV:Body/:/m:GetQuotationResponse/:/m:Quotation[@id=\"zs\"]/@name":{
         "algo":{
            "name":"mx.pfe.mask_string",
            "params":[
               "6"
            ],
            "guard":"output := input.x == 100 ",
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
            <m:Quotation id="zs" name="ZS_ATTR_TEXT_CONTEXT" >ZS_TEST_CONTEXT</m:Quotation>
            <m:Quotation id="ls" >LS</m:Quotation>
			<m:Code id="ww">
				<m:Name id="na">name</m:Name>
				<m:Age>18</m:Age>
			</m:Code>
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

func TestXMLShuffle_3(t *testing.T) {
	model := `{
	"namespaces":[
		{"SOAP-ENV":"http://www.w3.org/2001/12/soap-envelope"},
		{"m":"http://www.xyz.org/quotation"}
    ],
   "filters":{
      "denied":[
         {
            "match":"//SOAP-ENV:Envelope/:/SOAP-ENV:Body/:/m:GetQuotationResponse/:/m:Code/:/m:Name/@id",
            "guard":"na :=  xml(node,\".\",spaces) \n output := strEqualFold(na,\"na\") "
         }
      ],
      "rowfilter":{
         "//SOAP-ENV:Envelope/:/SOAP-ENV:Body/:/m:GetQuotationResponse/:/m:Quotation[@id=\"ls\"]":"output := true"
      }
   },
"shuffle":{
      "//SOAP-ENV:Envelope/:/SOAP-ENV:Body/:/m:GetQuotationResponse/:/m:Quotation[@id=\"zs\"]/@name":{
         "algo":{
            "name":"mx.pfe.mask_string",
            "params":[
               "6"
            ],
            "guard":"output := input.x == 100 ",
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
            <m:Quotation id="zs" name="ZS_ATTR_TEXT_CONTEXT" >ZS_TEST_CONTEXT</m:Quotation>
            <m:Quotation id="ls" >LS</m:Quotation>
			<m:Code id="ww">
				<m:Name id="na" sex="1" >name</m:Name>
				<m:Age>18</m:Age>
			</m:Code>
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

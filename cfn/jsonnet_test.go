package cfn_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/fujiwara/cfn-lookup/cfn"
	"github.com/google/go-jsonnet"
)

func TestJsonnetNativeFuncs(t *testing.T) {
	app := cfn.NewMock(&mockCfnClient{})
	ctx := context.Background()
	funcs := app.JsonnetNativeFuncs(ctx)
	if len(funcs) != 2 {
		t.Errorf("unexpected funcs: %v", funcs)
	}
	vm := jsonnet.MakeVM()
	for _, f := range funcs {
		vm.NativeFunction(f)
	}
	out, err := vm.EvaluateAnonymousSnippet("test.jsonnet", `
	local cfn_output = std.native('cfn_output');
	local cfn_export = std.native('cfn_export');
	{
		"output": cfn_output("test-stack", "test-key"),
		"export": cfn_export("test-export"),
	}`)
	if err != nil {
		t.Errorf("error: %v", err)
	}
	var result map[string]string
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Errorf("error: %v", err)
	}
	if result["output"] != "test-value" {
		t.Errorf("unexpected value: %s", result["output"])
	}
	if result["export"] != "test-export-value" {
		t.Errorf("unexpected value: %s", result["export"])
	}
}

package cfn_test

import (
	"context"
	"testing"

	"github.com/fujiwara/cfn-lookup/cfn"
)

func TestFuncMap(t *testing.T) {
	app := cfn.NewMock(&mockCfnClient{})
	ctx := context.Background()
	funcs := app.FuncMap(ctx)
	if len(funcs) != 2 {
		t.Errorf("unexpected funcs: %v", funcs)
	}
	fn := funcs["cfn_output"]
	vo := fn.(func(string, string) string)("test-stack", "test-key")
	if vo != "test-value" {
		t.Errorf("unexpected value: %s", vo)
	}
	fn = funcs["cfn_export"]
	ve := fn.(func(string) string)("test-export")
	if ve != "test-export-value" {
		t.Errorf("unexpected value: %s", ve)
	}
}

package cfn

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/google/go-jsonnet"
	"github.com/google/go-jsonnet/ast"
)

func JsonnetNativeFuncs(ctx context.Context, cfg aws.Config) ([]*jsonnet.NativeFunction, error) {
	cache := sync.Map{}
	app := New(cfg, &cache)
	return app.JsonnetNativeFuncs(ctx), nil
}

// JsonnetNativeFuncs provides native functions for jsonnet.
func (app *App) JsonnetNativeFuncs(ctx context.Context) []*jsonnet.NativeFunction {
	return []*jsonnet.NativeFunction{
		{
			Name:   "cfn_output",
			Params: []ast.Identifier{"stackName", "outputKey"},
			Func: func(p []interface{}) (interface{}, error) {
				stackName, ok := p[0].(string)
				if !ok {
					return nil, fmt.Errorf("stackName must be a string")
				}
				outputKey, ok := p[1].(string)
				if !ok {
					return nil, fmt.Errorf("outputKey must be a string")
				}
				return app.LookupOutput(ctx, stackName, outputKey)
			},
		},
		{
			Name:   "cfn_export",
			Params: []ast.Identifier{"name"},
			Func: func(p []interface{}) (interface{}, error) {
				name, ok := p[0].(string)
				if !ok {
					return nil, fmt.Errorf("name must be a string")
				}
				return app.LookupExport(ctx, name)
			},
		},
	}
}

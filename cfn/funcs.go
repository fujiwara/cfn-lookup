package cfn

import (
	"context"
	"fmt"
	"sync"
	"text/template"

	"github.com/aws/aws-sdk-go-v2/aws"
)

// FuncMap provides a tamplate.FuncMap. can lockup values from CFn stack.
func FuncMap(ctx context.Context, cfg aws.Config) (template.FuncMap, error) {
	cache := sync.Map{}
	app := New(cfg, &cache)
	return template.FuncMap{
		"cfn_output": func(stackName, outputKey string) string {
			value, err := app.LookupOutput(ctx, stackName, outputKey)
			if err != nil {
				panic(fmt.Sprintf("failed to lookup %s in stack %s: %s", outputKey, stackName, err))
			}
			return value
		},
		"cfn_export": func(name string) string {
			value, err := app.LookupExport(ctx, name)
			if err != nil {
				panic(fmt.Sprintf("failed to lookup %s in exports: %s", name, err))
			}
			return value
		},
	}, nil
}

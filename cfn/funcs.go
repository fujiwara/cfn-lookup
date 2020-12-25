package cfn

import (
	"fmt"
	"sync"
	"text/template"

	"github.com/aws/aws-sdk-go/aws/session"
)

// FuncMap provides a tamplate.FuncMap. can lockup values from CFn stack.
func FuncMap(sess *session.Session) (template.FuncMap, error) {
	cache := sync.Map{}
	app := New(sess, &cache)
	return template.FuncMap{
		"cfn_output": func(stackName, outputKey string) string {
			value, err := app.LookupOutput(stackName, outputKey)
			if err != nil {
				panic(fmt.Sprintf("failed to lookup %s in stack %s: %s", outputKey, stackName, err))
			}
			return value
		},
		"cfn_export": func(name string) string {
			value, err := app.LookupExport(name)
			if err != nil {
				panic(fmt.Sprintf("failed to lookup %s in exports: %s", name, err))
			}
			return value
		},
	}, nil
}

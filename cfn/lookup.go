package cfn

import (
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/pkg/errors"
)

// App represents an application
type App struct {
	cfn   *cloudformation.CloudFormation
	cache *sync.Map
}

// New creates an application instance
func New(sess *session.Session, cache *sync.Map) *App {
	return &App{
		cfn:   cloudformation.New(sess),
		cache: cache,
	}
}

// LookupOutput lookups output value for the stack.
func (a *App) LookupOutput(stackName, outputKey string) (outputValue string, err error) {
	if a.cache == nil {
		stack, err := getStack(a.cfn, stackName)
		if err != nil {
			return "", err
		}
		return lookupOutput(stack, outputKey)
	}

	key := "stack::" + stackName
	if s, found := a.cache.Load(key); found {
		// hit stack in cache
		return lookupOutput(s.(*cloudformation.Stack), outputKey)
	}

	stack, err := getStack(a.cfn, stackName)
	if err != nil {
		return "", err
	}
	a.cache.Store(key, stack)
	return lookupOutput(stack, outputKey)
}

func getStack(cfn *cloudformation.CloudFormation, stackName string) (*cloudformation.Stack, error) {
	out, err := cfn.DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to describe stacks %s", stackName)
	}
	if len(out.Stacks) == 0 {
		return nil, errors.Errorf("%s is not found", stackName)
	}
	return out.Stacks[0], nil
}

func lookupOutput(stack *cloudformation.Stack, outputKey string) (outputValue string, err error) {
	for _, output := range stack.Outputs {
		if aws.StringValue(output.OutputKey) == outputKey {
			return aws.StringValue(output.OutputValue), nil
		}
	}
	return "", errors.Errorf("outputKey %s is not found in stack %s", outputKey, *stack.StackName)
}

// LookupExport lookups exported value.
func (a *App) LookupExport(name string) (value string, err error) {
	if a.cache == nil {
		res, err := getExports(a.cfn)
		if err != nil {
			return "", err
		}
		return lookupExport(res, name)
	}
	key := "export::"
	if e, found := a.cache.Load(key); found {
		return lookupExport(e.(*cloudformation.ListExportsOutput), name)
	}

	res, err := getExports(a.cfn)
	if err != nil {
		return "", err
	}
	a.cache.Store(key, res)
	return lookupExport(res, name)
}

// ExportedNames lists names of exports.
func (a *App) ExportedNames() ([]string, error) {
	if a.cache == nil {
		res, err := getExports(a.cfn)
		if err != nil {
			return nil, err
		}
		return listNames(res)
	}
	key := "export::"
	if e, found := a.cache.Load(key); found {
		return listNames(e.(*cloudformation.ListExportsOutput))
	}

	res, err := getExports(a.cfn)
	if err != nil {
		return nil, err
	}
	a.cache.Store(key, res)
	return listNames(res)
}

func getExports(cfn *cloudformation.CloudFormation) (*cloudformation.ListExportsOutput, error) {
	var nextToken *string
	res := &cloudformation.ListExportsOutput{}
	for {
		out, err := cfn.ListExports(&cloudformation.ListExportsInput{
			NextToken: nextToken,
		})
		if err != nil {
			return nil, errors.Wrapf(err, "failed to list exports")
		}
		for _, export := range out.Exports {
			export := export
			res.Exports = append(res.Exports, export)
		}
		if nextToken = out.NextToken; nextToken == nil {
			break
		}
	}
	return res, nil
}

func lookupExport(res *cloudformation.ListExportsOutput, name string) (string, error) {
	for _, ex := range res.Exports {
		if aws.StringValue(ex.Name) == name {
			return aws.StringValue(ex.Value), nil
		}
	}
	return "", errors.Errorf("%s is not found in exports", name)
}

func listNames(res *cloudformation.ListExportsOutput) ([]string, error) {
	names := make([]string, 0, len(res.Exports))
	for _, ex := range res.Exports {
		names = append(names, *ex.Name)
	}
	return names, nil
}

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

type stack = cloudformation.Stack
type export = cloudformation.Export

// New creates an application instance
func New(sess *session.Session, cache *sync.Map) *App {
	return &App{
		cfn:   cloudformation.New(sess),
		cache: cache,
	}
}

// LookupOutput lookups output value for the stack.
func (a *App) LookupOutput(stackName, outputKey string) (outputValue string, err error) {
	stack, err := getStackWithCache(a.cfn, stackName, a.cache)
	if err != nil {
		return "", err
	}
	return lookupOutput(stack, outputKey)
}

func getStackWithCache(cfn *cloudformation.CloudFormation, stackName string, cache *sync.Map) (*stack, error) {
	if cache == nil {
		return getStack(cfn, stackName)
	}

	key := "stack::" + stackName
	if s, found := cache.Load(key); found {
		return s.(*stack), nil
	}

	if s, err := getStack(cfn, stackName); err != nil {
		return nil, err
	} else {
		cache.Store(key, s)
		return s, nil
	}
}

func getStack(cfn *cloudformation.CloudFormation, stackName string) (*stack, error) {
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

func lookupOutput(stack *stack, outputKey string) (outputValue string, err error) {
	for _, output := range stack.Outputs {
		if aws.StringValue(output.OutputKey) == outputKey {
			return aws.StringValue(output.OutputValue), nil
		}
	}
	return "", errors.Errorf("outputKey %s is not found in stack %s", outputKey, *stack.StackName)
}

// LookupExport lookups exported value.
func (a *App) LookupExport(name string) (value string, err error) {
	ex, err := getExportsWithCache(a.cfn, a.cache)
	if err != nil {
		return "", err
	}
	return lookupExport(ex, name)
}

// ExportedNames lists names of exports.
func (a *App) ExportedNames() ([]string, error) {
	ex, err := getExportsWithCache(a.cfn, a.cache)
	if err != nil {
		return nil, err
	}
	return listNames(ex)
}

func getExportsWithCache(cfn *cloudformation.CloudFormation, cache *sync.Map) ([]*export, error) {
	if cache == nil {
		return getExports(cfn)
	}

	key := "export::"
	if e, found := cache.Load(key); found {
		return e.([]*export), nil
	}

	if ex, err := getExports(cfn); err != nil {
		return nil, err
	} else {
		cache.Store(key, ex)
		return ex, nil
	}
}

func getExports(cfn *cloudformation.CloudFormation) ([]*export, error) {
	var nextToken *string
	exs := make([]*export, 0)
	for {
		out, err := cfn.ListExports(&cloudformation.ListExportsInput{
			NextToken: nextToken,
		})
		if err != nil {
			return nil, errors.Wrapf(err, "failed to list exports")
		}
		for _, ex := range out.Exports {
			ex := ex
			exs = append(exs, ex)
		}
		if nextToken = out.NextToken; nextToken == nil {
			break
		}
	}
	return exs, nil
}

func lookupExport(exs []*export, name string) (string, error) {
	for _, ex := range exs {
		if aws.StringValue(ex.Name) == name {
			return aws.StringValue(ex.Value), nil
		}
	}
	return "", errors.Errorf("%s is not found in exports", name)
}

func listNames(exs []*export) ([]string, error) {
	names := make([]string, 0, len(exs))
	for _, ex := range exs {
		names = append(names, aws.StringValue(ex.Name))
	}
	return names, nil
}

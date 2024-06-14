package cfn

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	types "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
)

// App represents an application
type App struct {
	cfn   cfnClient
	cache *sync.Map
}

type cfnClient interface {
	DescribeStacks(ctx context.Context, params *cloudformation.DescribeStacksInput, optFns ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error)
	ListExports(ctx context.Context, params *cloudformation.ListExportsInput, optFns ...func(*cloudformation.Options)) (*cloudformation.ListExportsOutput, error)
}

type stack = types.Stack
type export = types.Export

// New creates an application instance
func New(cfg aws.Config, cache *sync.Map) *App {
	return &App{
		cfn:   cloudformation.NewFromConfig(cfg),
		cache: cache,
	}
}

// LookupOutput lookups output value for the stack.
func (a *App) LookupOutput(ctx context.Context, stackName, outputKey string) (outputValue string, err error) {
	stack, err := getStackWithCache(ctx, a.cfn, stackName, a.cache)
	if err != nil {
		return "", err
	}
	return lookupOutput(stack, outputKey)
}

// ListOutput lists output keys for the stack.
func (a *App) ListOutput(ctx context.Context, stackName string) ([]string, error) {
	stack, err := getStackWithCache(ctx, a.cfn, stackName, a.cache)
	if err != nil {
		return nil, err
	}
	return listOutput(stack)
}

func getStackWithCache(ctx context.Context, cfn cfnClient, stackName string, cache *sync.Map) (*stack, error) {
	if cache == nil {
		return getStack(ctx, cfn, stackName)
	}

	key := "stack::" + stackName
	if s, found := cache.Load(key); found {
		return s.(*stack), nil
	}

	if s, err := getStack(ctx, cfn, stackName); err != nil {
		return nil, err
	} else {
		cache.Store(key, s)
		return s, nil
	}
}

func getStack(ctx context.Context, cfn cfnClient, stackName string) (*stack, error) {
	out, err := cfn.DescribeStacks(ctx, &cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe stacks %s: %w", stackName, err)
	}
	if len(out.Stacks) == 0 {
		return nil, fmt.Errorf("%s is not found: %w", stackName, err)
	}
	return &out.Stacks[0], nil
}

func lookupOutput(stack *stack, outputKey string) (outputValue string, err error) {
	for _, output := range stack.Outputs {
		if aws.ToString(output.OutputKey) == outputKey {
			return aws.ToString(output.OutputValue), nil
		}
	}
	return "", fmt.Errorf("outputKey %s is not found in stack %s", outputKey, *stack.StackName)
}

func listOutput(stack *stack) (keys []string, err error) {
	for _, output := range stack.Outputs {
		keys = append(keys, aws.ToString(output.OutputKey))
	}
	return
}

// LookupExport lookups exported value.
func (a *App) LookupExport(ctx context.Context, name string) (value string, err error) {
	ex, err := getExportsWithCache(ctx, a.cfn, a.cache)
	if err != nil {
		return "", err
	}
	return lookupExport(ex, name)
}

// ExportedNames lists names of exports.
func (a *App) ExportedNames(ctx context.Context) ([]string, error) {
	ex, err := getExportsWithCache(ctx, a.cfn, a.cache)
	if err != nil {
		return nil, err
	}
	return listNames(ex)
}

func getExportsWithCache(ctx context.Context, cfn cfnClient, cache *sync.Map) ([]*export, error) {
	if cache == nil {
		return getExports(ctx, cfn)
	}

	key := "export::"
	if e, found := cache.Load(key); found {
		return e.([]*export), nil
	}

	if ex, err := getExports(ctx, cfn); err != nil {
		return nil, err
	} else {
		cache.Store(key, ex)
		return ex, nil
	}
}

func getExports(ctx context.Context, cfn cfnClient) ([]*export, error) {
	var nextToken *string
	exs := make([]*export, 0)
	for {
		out, err := cfn.ListExports(ctx, &cloudformation.ListExportsInput{
			NextToken: nextToken,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list exports: %w", err)
		}
		for _, ex := range out.Exports {
			ex := ex
			exs = append(exs, &ex)
		}
		if nextToken = out.NextToken; nextToken == nil {
			break
		}
	}
	return exs, nil
}

func lookupExport(exs []*export, name string) (string, error) {
	for _, ex := range exs {
		if aws.ToString(ex.Name) == name {
			return aws.ToString(ex.Value), nil
		}
	}
	return "", fmt.Errorf("%s is not found in exports", name)
}

func listNames(exs []*export) ([]string, error) {
	names := make([]string, 0, len(exs))
	for _, ex := range exs {
		names = append(names, aws.ToString(ex.Name))
	}
	return names, nil
}

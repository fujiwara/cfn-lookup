package cfn_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	types "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/fujiwara/cfn-lookup/cfn"
)

type mockCfnClient struct{}

func (c *mockCfnClient) DescribeStacks(ctx context.Context, params *cloudformation.DescribeStacksInput, optFns ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error) {
	return &cloudformation.DescribeStacksOutput{
		Stacks: []types.Stack{
			{
				StackName: aws.String("test-stack"),
				StackId:   aws.String("test-stack-id"),
				Outputs: []types.Output{
					{
						OutputKey:   aws.String("test-key"),
						OutputValue: aws.String("test-value"),
					},
				},
			},
		},
	}, nil
}

func (c *mockCfnClient) ListExports(ctx context.Context, params *cloudformation.ListExportsInput, optFns ...func(*cloudformation.Options)) (*cloudformation.ListExportsOutput, error) {
	return &cloudformation.ListExportsOutput{
		Exports: []types.Export{
			{
				Name:  aws.String("test-export"),
				Value: aws.String("test-export-value"),
			},
		},
	}, nil
}

func TestLookup(t *testing.T) {
	app := cfn.NewMock(&mockCfnClient{})
	ctx := context.Background()
	vo, err := app.LookupOutput(ctx, "test-stack", "test-key")
	if err != nil {
		t.Errorf("error: %v", err)
	}
	if vo != "test-value" {
		t.Errorf("unexpected value: %s", vo)
	}

	ve, err := app.LookupExport(ctx, "test-export")
	if err != nil {
		t.Errorf("error: %v", err)
	}
	if ve != "test-export-value" {
		t.Errorf("unexpected value: %s", ve)
	}
}

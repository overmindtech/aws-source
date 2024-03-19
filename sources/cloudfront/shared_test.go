package cloudfront

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/overmindtech/aws-source/sources"
)

func (c TestCloudFrontClient) ListTagsForResource(ctx context.Context, params *cloudfront.ListTagsForResourceInput, optFns ...func(*cloudfront.Options)) (*cloudfront.ListTagsForResourceOutput, error) {
	return &cloudfront.ListTagsForResourceOutput{
		Tags: &types.Tags{
			Items: []types.Tag{
				{
					Key:   sources.PtrString("foo"),
					Value: sources.PtrString("bar"),
				},
			},
		},
	}, nil
}

type TestCloudFrontClient struct{}

func GetAutoConfig(t *testing.T) (*cloudfront.Client, string, string) {
	config, account, region := sources.GetAutoConfig(t)
	client := cloudfront.NewFromConfig(config)

	return client, account, region
}

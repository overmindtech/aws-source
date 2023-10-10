package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func regionInputMapperGet(scope string, query string) (*ec2.DescribeRegionsInput, error) {
	return &ec2.DescribeRegionsInput{
		RegionNames: []string{
			query,
		},
	}, nil
}

func regionInputMapperList(scope string) (*ec2.DescribeRegionsInput, error) {
	return &ec2.DescribeRegionsInput{}, nil
}

func regionOutputMapper(_ context.Context, _ *ec2.Client, scope string, _ *ec2.DescribeRegionsInput, output *ec2.DescribeRegionsOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, ni := range output.Regions {
		var err error
		var attrs *sdp.ItemAttributes
		attrs, err = sources.ToAttributesCase(ni)

		if err != nil {
			return nil, &sdp.QueryError{
				ErrorType:   sdp.QueryError_OTHER,
				ErrorString: err.Error(),
				Scope:       scope,
			}
		}

		item := sdp.Item{
			Type:            "ec2-region",
			UniqueAttribute: "regionName",
			Scope:           scope,
			Attributes:      attrs,
		}

		items = append(items, &item)
	}

	return items, nil
}

//go:generate docgen ../../docs-data
// +overmind:type ec2-region
// +overmind:descriptiveType Region
// +overmind:get Get a region by name
// +overmind:list List all regions
// +overmind:group AWS

func NewRegionSource(config aws.Config, accountID string, limit *sources.LimitBucket) *sources.DescribeOnlySource[*ec2.DescribeRegionsInput, *ec2.DescribeRegionsOutput, *ec2.Client, *ec2.Options] {
	return &sources.DescribeOnlySource[*ec2.DescribeRegionsInput, *ec2.DescribeRegionsOutput, *ec2.Client, *ec2.Options]{
		Config:    config,
		Client:    ec2.NewFromConfig(config),
		AccountID: accountID,
		ItemType:  "ec2-region",
		DescribeFunc: func(ctx context.Context, client *ec2.Client, input *ec2.DescribeRegionsInput) (*ec2.DescribeRegionsOutput, error) {
			limit.Wait(ctx) // Wait for rate limiting // Wait for late limiting
			return client.DescribeRegions(ctx, input)
		},
		InputMapperGet:  regionInputMapperGet,
		InputMapperList: regionInputMapperList,
		OutputMapper:    regionOutputMapper,
	}
}

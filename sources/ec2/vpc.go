package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func vpcInputMapperGet(scope string, query string) (*ec2.DescribeVpcsInput, error) {
	return &ec2.DescribeVpcsInput{
		VpcIds: []string{
			query,
		},
	}, nil
}

func vpcInputMapperList(scope string) (*ec2.DescribeVpcsInput, error) {
	return &ec2.DescribeVpcsInput{}, nil
}

func vpcOutputMapper(_ context.Context, _ *ec2.Client, scope string, _ *ec2.DescribeVpcsInput, output *ec2.DescribeVpcsOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, vpc := range output.Vpcs {
		var err error
		var attrs *sdp.ItemAttributes
		attrs, err = sources.ToAttributesCase(vpc)

		if err != nil {
			return nil, &sdp.QueryError{
				ErrorType:   sdp.QueryError_OTHER,
				ErrorString: err.Error(),
				Scope:       scope,
			}
		}

		item := sdp.Item{
			Type:            "ec2-vpc",
			UniqueAttribute: "vpcId",
			Scope:           scope,
			Attributes:      attrs,
			Tags:            tagsToMap(vpc.Tags),
		}

		items = append(items, &item)
	}

	return items, nil
}

//go:generate docgen ../../docs-data
// +overmind:type ec2-vpc
// +overmind:descriptiveType VPC
// +overmind:get Get a VPC by ID
// +overmind:list List all VPCs
// +overmind:group AWS
// +overmind:terraform:queryMap aws_vpc.id

func NewVpcSource(config aws.Config, accountID string, limit *sources.LimitBucket) *sources.DescribeOnlySource[*ec2.DescribeVpcsInput, *ec2.DescribeVpcsOutput, *ec2.Client, *ec2.Options] {
	return &sources.DescribeOnlySource[*ec2.DescribeVpcsInput, *ec2.DescribeVpcsOutput, *ec2.Client, *ec2.Options]{
		Config:    config,
		Client:    ec2.NewFromConfig(config),
		AccountID: accountID,
		ItemType:  "ec2-vpc",
		DescribeFunc: func(ctx context.Context, client *ec2.Client, input *ec2.DescribeVpcsInput) (*ec2.DescribeVpcsOutput, error) {
			limit.Wait(ctx) // Wait for rate limiting // Wait for late limiting
			return client.DescribeVpcs(ctx, input)
		},
		InputMapperGet:  vpcInputMapperGet,
		InputMapperList: vpcInputMapperList,
		PaginatorBuilder: func(client *ec2.Client, params *ec2.DescribeVpcsInput) sources.Paginator[*ec2.DescribeVpcsOutput, *ec2.Options] {
			return ec2.NewDescribeVpcsPaginator(client, params)
		},
		OutputMapper: vpcOutputMapper,
	}
}

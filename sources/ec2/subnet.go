package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func subnetInputMapperGet(scope string, query string) (*ec2.DescribeSubnetsInput, error) {
	return &ec2.DescribeSubnetsInput{
		SubnetIds: []string{
			query,
		},
	}, nil
}

func subnetInputMapperList(scope string) (*ec2.DescribeSubnetsInput, error) {
	return &ec2.DescribeSubnetsInput{}, nil
}

func subnetOutputMapper(scope string, _ *ec2.DescribeSubnetsInput, output *ec2.DescribeSubnetsOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, subnet := range output.Subnets {
		var err error
		var attrs *sdp.ItemAttributes
		attrs, err = sources.ToAttributesCase(subnet)

		if err != nil {
			return nil, &sdp.QueryError{
				ErrorType:   sdp.QueryError_OTHER,
				ErrorString: err.Error(),
				Scope:       scope,
			}
		}

		item := sdp.Item{
			Type:            "ec2-subnet",
			UniqueAttribute: "subnetId",
			Scope:           scope,
			Attributes:      attrs,
		}

		if subnet.AvailabilityZone != nil {
			// +overmind:link ec2-availability-zone
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
				Type:   "ec2-availability-zone",
				Method: sdp.QueryMethod_GET,
				Query:  *subnet.AvailabilityZone,
				Scope:  scope,
			})
		}

		if subnet.VpcId != nil {
			// +overmind:link ec2-vpc
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
				Type:   "ec2-vpc",
				Method: sdp.QueryMethod_GET,
				Query:  *subnet.VpcId,
				Scope:  scope,
			})
		}

		items = append(items, &item)
	}

	return items, nil
}

//go:generate docgen ../../docs-data
// +overmind:type ec2-subnet
// +overmind:descriptiveType EC2 Subnet
// +overmind:get Get a subnet by ID
// +overmind:list List all subnets
// +overmind:search Search for subnets by ARN
// +overmind:group AWS

func NewSubnetSource(config aws.Config, accountID string, limit *LimitBucket) *sources.DescribeOnlySource[*ec2.DescribeSubnetsInput, *ec2.DescribeSubnetsOutput, *ec2.Client, *ec2.Options] {
	return &sources.DescribeOnlySource[*ec2.DescribeSubnetsInput, *ec2.DescribeSubnetsOutput, *ec2.Client, *ec2.Options]{
		Config:    config,
		Client:    ec2.NewFromConfig(config),
		AccountID: accountID,
		ItemType:  "ec2-subnet",
		DescribeFunc: func(ctx context.Context, client *ec2.Client, input *ec2.DescribeSubnetsInput) (*ec2.DescribeSubnetsOutput, error) {
			<-limit.C // Wait for late limiting
			return client.DescribeSubnets(ctx, input)
		},
		InputMapperGet:  subnetInputMapperGet,
		InputMapperList: subnetInputMapperList,
		PaginatorBuilder: func(client *ec2.Client, params *ec2.DescribeSubnetsInput) sources.Paginator[*ec2.DescribeSubnetsOutput, *ec2.Options] {
			return ec2.NewDescribeSubnetsPaginator(client, params)
		},
		OutputMapper: subnetOutputMapper,
	}
}

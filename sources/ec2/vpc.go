package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func VpcInputMapperGet(scope string, query string) (*ec2.DescribeVpcsInput, error) {
	return &ec2.DescribeVpcsInput{
		VpcIds: []string{
			query,
		},
	}, nil
}

func VpcInputMapperList(scope string) (*ec2.DescribeVpcsInput, error) {
	return &ec2.DescribeVpcsInput{}, nil
}

func VpcOutputMapper(scope string, output *ec2.DescribeVpcsOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, vpc := range output.Vpcs {
		var err error
		var attrs *sdp.ItemAttributes
		attrs, err = sources.ToAttributesCase(vpc)

		if err != nil {
			return nil, &sdp.ItemRequestError{
				ErrorType:   sdp.ItemRequestError_OTHER,
				ErrorString: err.Error(),
				Scope:       scope,
			}
		}

		item := sdp.Item{
			Type:            "ec2-vpc",
			UniqueAttribute: "vpcId",
			Scope:           scope,
			Attributes:      attrs,
		}

		items = append(items, &item)
	}

	return items, nil
}

func NewVpcSource(config aws.Config, accountID string, limit *LimitBucket) *sources.DescribeOnlySource[*ec2.DescribeVpcsInput, *ec2.DescribeVpcsOutput, *ec2.Client, *ec2.Options] {
	return &sources.DescribeOnlySource[*ec2.DescribeVpcsInput, *ec2.DescribeVpcsOutput, *ec2.Client, *ec2.Options]{
		Config:    config,
		AccountID: accountID,
		ItemType:  "ec2-vpc",
		DescribeFunc: func(ctx context.Context, client *ec2.Client, input *ec2.DescribeVpcsInput) (*ec2.DescribeVpcsOutput, error) {
			<-limit.C // Wait for late limiting
			return client.DescribeVpcs(ctx, input)
		},
		InputMapperGet:  VpcInputMapperGet,
		InputMapperList: VpcInputMapperList,
		PaginatorBuilder: func(client *ec2.Client, params *ec2.DescribeVpcsInput) sources.Paginator[*ec2.DescribeVpcsOutput, *ec2.Options] {
			return ec2.NewDescribeVpcsPaginator(client, params)
		},
		OutputMapper: VpcOutputMapper,
	}
}

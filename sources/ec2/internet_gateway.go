package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func internetGatewayInputMapperGet(scope string, query string) (*ec2.DescribeInternetGatewaysInput, error) {
	return &ec2.DescribeInternetGatewaysInput{
		InternetGatewayIds: []string{
			query,
		},
	}, nil
}

func internetGatewayInputMapperList(scope string) (*ec2.DescribeInternetGatewaysInput, error) {
	return &ec2.DescribeInternetGatewaysInput{}, nil
}

func internetGatewayOutputMapper(scope string, _ *ec2.DescribeInternetGatewaysInput, output *ec2.DescribeInternetGatewaysOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, gw := range output.InternetGateways {
		var err error
		var attrs *sdp.ItemAttributes
		attrs, err = sources.ToAttributesCase(gw)

		if err != nil {
			return nil, &sdp.QueryError{
				ErrorType:   sdp.QueryError_OTHER,
				ErrorString: err.Error(),
				Scope:       scope,
			}
		}

		item := sdp.Item{
			Type:            "ec2-internet-gateway",
			UniqueAttribute: "internetGatewayId",
			Scope:           scope,
			Attributes:      attrs,
		}

		// VPCs
		for _, attachment := range gw.Attachments {
			if attachment.VpcId != nil {
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
					Type:   "ec2-vpc",
					Method: sdp.QueryMethod_GET,
					Query:  *attachment.VpcId,
					Scope:  scope,
				})
			}
		}

		items = append(items, &item)
	}

	return items, nil
}

func NewInternetGatewaySource(config aws.Config, accountID string, limit *LimitBucket) *sources.DescribeOnlySource[*ec2.DescribeInternetGatewaysInput, *ec2.DescribeInternetGatewaysOutput, *ec2.Client, *ec2.Options] {
	return &sources.DescribeOnlySource[*ec2.DescribeInternetGatewaysInput, *ec2.DescribeInternetGatewaysOutput, *ec2.Client, *ec2.Options]{
		Config:    config,
		Client:    ec2.NewFromConfig(config),
		AccountID: accountID,
		ItemType:  "ec2-internet-gateway",
		DescribeFunc: func(ctx context.Context, client *ec2.Client, input *ec2.DescribeInternetGatewaysInput) (*ec2.DescribeInternetGatewaysOutput, error) {
			<-limit.C // Wait for late limiting
			return client.DescribeInternetGateways(ctx, input)
		},
		InputMapperGet:  internetGatewayInputMapperGet,
		InputMapperList: internetGatewayInputMapperList,
		PaginatorBuilder: func(client *ec2.Client, params *ec2.DescribeInternetGatewaysInput) sources.Paginator[*ec2.DescribeInternetGatewaysOutput, *ec2.Options] {
			return ec2.NewDescribeInternetGatewaysPaginator(client, params)
		},
		OutputMapper: internetGatewayOutputMapper,
	}
}

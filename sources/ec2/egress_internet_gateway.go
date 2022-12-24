package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func EgressOnlyInternetGatewayInputMapperGet(scope string, query string) (*ec2.DescribeEgressOnlyInternetGatewaysInput, error) {
	return &ec2.DescribeEgressOnlyInternetGatewaysInput{
		EgressOnlyInternetGatewayIds: []string{
			query,
		},
	}, nil
}

func EgressOnlyInternetGatewayInputMapperList(scope string) (*ec2.DescribeEgressOnlyInternetGatewaysInput, error) {
	return &ec2.DescribeEgressOnlyInternetGatewaysInput{}, nil
}

func EgressOnlyInternetGatewayOutputMapper(scope string, output *ec2.DescribeEgressOnlyInternetGatewaysOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, gw := range output.EgressOnlyInternetGateways {
		var err error
		var attrs *sdp.ItemAttributes
		attrs, err = sources.ToAttributesCase(gw)

		if err != nil {
			return nil, &sdp.ItemRequestError{
				ErrorType:   sdp.ItemRequestError_OTHER,
				ErrorString: err.Error(),
				Scope:       scope,
			}
		}

		item := sdp.Item{
			Type:            "ec2-egress-only-internet-gateway",
			UniqueAttribute: "egressOnlyInternetGatewayId",
			Scope:           scope,
			Attributes:      attrs,
		}

		for _, attachment := range gw.Attachments {
			if attachment.VpcId != nil {
				item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
					Type:   "ec2-vpc",
					Method: sdp.RequestMethod_GET,
					Query:  *attachment.VpcId,
					Scope:  scope,
				})
			}
		}

		items = append(items, &item)
	}

	return items, nil
}

func NewEgressOnlyInternetGatewaySource(config aws.Config, accountID string) *sources.AWSSource[*ec2.DescribeEgressOnlyInternetGatewaysInput, *ec2.DescribeEgressOnlyInternetGatewaysOutput, *ec2.Client, *ec2.Options] {
	return &sources.AWSSource[*ec2.DescribeEgressOnlyInternetGatewaysInput, *ec2.DescribeEgressOnlyInternetGatewaysOutput, *ec2.Client, *ec2.Options]{
		Config:    config,
		AccountID: accountID,
		ItemType:  "ec2-EgressOnlyInternetGateway",
		DescribeFunc: func(ctx context.Context, client *ec2.Client, input *ec2.DescribeEgressOnlyInternetGatewaysInput) (*ec2.DescribeEgressOnlyInternetGatewaysOutput, error) {
			return client.DescribeEgressOnlyInternetGateways(ctx, input)
		},
		InputMapperGet:  EgressOnlyInternetGatewayInputMapperGet,
		InputMapperList: EgressOnlyInternetGatewayInputMapperList,
		PaginatorBuilder: func(client *ec2.Client, params *ec2.DescribeEgressOnlyInternetGatewaysInput) sources.Paginator[*ec2.DescribeEgressOnlyInternetGatewaysOutput, *ec2.Options] {
			return ec2.NewDescribeEgressOnlyInternetGatewaysPaginator(client, params)
		},
		OutputMapper: EgressOnlyInternetGatewayOutputMapper,
	}
}

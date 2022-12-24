package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func NetworkAclInputMapperGet(scope string, query string) (*ec2.DescribeNetworkAclsInput, error) {
	return &ec2.DescribeNetworkAclsInput{
		NetworkAclIds: []string{
			query,
		},
	}, nil
}

func NetworkAclInputMapperList(scope string) (*ec2.DescribeNetworkAclsInput, error) {
	return &ec2.DescribeNetworkAclsInput{}, nil
}

func NetworkAclOutputMapper(scope string, output *ec2.DescribeNetworkAclsOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, networkAcl := range output.NetworkAcls {
		var err error
		var attrs *sdp.ItemAttributes
		attrs, err = sources.ToAttributesCase(networkAcl)

		if err != nil {
			return nil, &sdp.ItemRequestError{
				ErrorType:   sdp.ItemRequestError_OTHER,
				ErrorString: err.Error(),
				Scope:       scope,
			}
		}

		item := sdp.Item{
			Type:            "ec2-network-acl",
			UniqueAttribute: "networkAclId",
			Scope:           scope,
			Attributes:      attrs,
		}

		for _, assoc := range networkAcl.Associations {
			if assoc.SubnetId != nil {
				item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
					Type:   "ec2-subnet",
					Method: sdp.RequestMethod_GET,
					Query:  *assoc.SubnetId,
					Scope:  scope,
				})
			}
		}

		if networkAcl.VpcId != nil {
			item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
				Type:   "ec2-vpc",
				Method: sdp.RequestMethod_GET,
				Query:  *networkAcl.VpcId,
				Scope:  scope,
			})
		}

		items = append(items, &item)
	}

	return items, nil
}

func NewNetworkAclSource(config aws.Config, accountID string) *sources.AWSSource[*ec2.DescribeNetworkAclsInput, *ec2.DescribeNetworkAclsOutput, *ec2.Client, *ec2.Options] {
	return &sources.AWSSource[*ec2.DescribeNetworkAclsInput, *ec2.DescribeNetworkAclsOutput, *ec2.Client, *ec2.Options]{
		Config:    config,
		AccountID: accountID,
		ItemType:  "ec2-network-acl",
		DescribeFunc: func(ctx context.Context, client *ec2.Client, input *ec2.DescribeNetworkAclsInput) (*ec2.DescribeNetworkAclsOutput, error) {
			return client.DescribeNetworkAcls(ctx, input)
		},
		InputMapperGet:  NetworkAclInputMapperGet,
		InputMapperList: NetworkAclInputMapperList,
		PaginatorBuilder: func(client *ec2.Client, params *ec2.DescribeNetworkAclsInput) sources.Paginator[*ec2.DescribeNetworkAclsOutput, *ec2.Options] {
			return ec2.NewDescribeNetworkAclsPaginator(client, params)
		},
		OutputMapper: NetworkAclOutputMapper,
	}
}

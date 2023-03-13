package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func NetworkInterfacePermissionInputMapperGet(scope string, query string) (*ec2.DescribeNetworkInterfacePermissionsInput, error) {
	return &ec2.DescribeNetworkInterfacePermissionsInput{
		NetworkInterfacePermissionIds: []string{
			query,
		},
	}, nil
}

func NetworkInterfacePermissionInputMapperList(scope string) (*ec2.DescribeNetworkInterfacePermissionsInput, error) {
	return &ec2.DescribeNetworkInterfacePermissionsInput{}, nil
}

func NetworkInterfacePermissionOutputMapper(scope string, _ *ec2.DescribeNetworkInterfacePermissionsInput, output *ec2.DescribeNetworkInterfacePermissionsOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, ni := range output.NetworkInterfacePermissions {
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
			Type:            "ec2-network-interface-permission",
			UniqueAttribute: "networkInterfacePermissionId",
			Scope:           scope,
			Attributes:      attrs,
		}

		if ni.NetworkInterfaceId != nil {
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
				Type:   "ec2-network-interface",
				Method: sdp.QueryMethod_GET,
				Query:  *ni.NetworkInterfaceId,
				Scope:  scope,
			})
		}

		items = append(items, &item)
	}

	return items, nil
}

func NewNetworkInterfacePermissionSource(config aws.Config, accountID string, limit *LimitBucket) *sources.DescribeOnlySource[*ec2.DescribeNetworkInterfacePermissionsInput, *ec2.DescribeNetworkInterfacePermissionsOutput, *ec2.Client, *ec2.Options] {
	return &sources.DescribeOnlySource[*ec2.DescribeNetworkInterfacePermissionsInput, *ec2.DescribeNetworkInterfacePermissionsOutput, *ec2.Client, *ec2.Options]{
		Config:    config,
		Client:    ec2.NewFromConfig(config),
		AccountID: accountID,
		ItemType:  "ec2-network-interface-permission",
		DescribeFunc: func(ctx context.Context, client *ec2.Client, input *ec2.DescribeNetworkInterfacePermissionsInput) (*ec2.DescribeNetworkInterfacePermissionsOutput, error) {
			<-limit.C // Wait for late limiting
			return client.DescribeNetworkInterfacePermissions(ctx, input)
		},
		InputMapperGet:  NetworkInterfacePermissionInputMapperGet,
		InputMapperList: NetworkInterfacePermissionInputMapperList,
		PaginatorBuilder: func(client *ec2.Client, params *ec2.DescribeNetworkInterfacePermissionsInput) sources.Paginator[*ec2.DescribeNetworkInterfacePermissionsOutput, *ec2.Options] {
			return ec2.NewDescribeNetworkInterfacePermissionsPaginator(client, params)
		},
		OutputMapper: NetworkInterfacePermissionOutputMapper,
	}
}

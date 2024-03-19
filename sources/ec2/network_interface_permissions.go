package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func networkInterfacePermissionInputMapperGet(scope string, query string) (*ec2.DescribeNetworkInterfacePermissionsInput, error) {
	return &ec2.DescribeNetworkInterfacePermissionsInput{
		NetworkInterfacePermissionIds: []string{
			query,
		},
	}, nil
}

func networkInterfacePermissionInputMapperList(scope string) (*ec2.DescribeNetworkInterfacePermissionsInput, error) {
	return &ec2.DescribeNetworkInterfacePermissionsInput{}, nil
}

func networkInterfacePermissionOutputMapper(_ context.Context, _ *ec2.Client, scope string, _ *ec2.DescribeNetworkInterfacePermissionsInput, output *ec2.DescribeNetworkInterfacePermissionsOutput) ([]*sdp.Item, error) {
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
			// +overmind:link ec2-network-interface
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "ec2-network-interface",
					Method: sdp.QueryMethod_GET,
					Query:  *ni.NetworkInterfaceId,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					// These permissions are tightly linked
					In:  true,
					Out: true,
				},
			})
		}

		items = append(items, &item)
	}

	return items, nil
}

//go:generate docgen ../../docs-data
// +overmind:type ec2-network-interface-permission
// +overmind:descriptiveType Network Interface Permission
// +overmind:get Get a network interface permission by ID
// +overmind:list List all network interface permissions
// +overmind:search Search network interface permissions by ARN
// +overmind:group AWS

func NewNetworkInterfacePermissionSource(client *ec2.Client, accountID string, region string, limit *sources.LimitBucket) *sources.DescribeOnlySource[*ec2.DescribeNetworkInterfacePermissionsInput, *ec2.DescribeNetworkInterfacePermissionsOutput, *ec2.Client, *ec2.Options] {
	return &sources.DescribeOnlySource[*ec2.DescribeNetworkInterfacePermissionsInput, *ec2.DescribeNetworkInterfacePermissionsOutput, *ec2.Client, *ec2.Options]{

		Client:    client,
		AccountID: accountID,
		ItemType:  "ec2-network-interface-permission",
		DescribeFunc: func(ctx context.Context, client *ec2.Client, input *ec2.DescribeNetworkInterfacePermissionsInput) (*ec2.DescribeNetworkInterfacePermissionsOutput, error) {
			limit.Wait(ctx) // Wait for rate limiting // Wait for late limiting
			return client.DescribeNetworkInterfacePermissions(ctx, input)
		},
		InputMapperGet:  networkInterfacePermissionInputMapperGet,
		InputMapperList: networkInterfacePermissionInputMapperList,
		PaginatorBuilder: func(client *ec2.Client, params *ec2.DescribeNetworkInterfacePermissionsInput) sources.Paginator[*ec2.DescribeNetworkInterfacePermissionsOutput, *ec2.Options] {
			return ec2.NewDescribeNetworkInterfacePermissionsPaginator(client, params)
		},
		OutputMapper: networkInterfacePermissionOutputMapper,
	}
}

package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"

	"github.com/overmindtech/aws-source/adapterhelpers"
	"github.com/overmindtech/sdp-go"
)

func vpcPeeringConnectionOutputMapper(_ context.Context, _ *ec2.Client, scope string, input *ec2.DescribeVpcPeeringConnectionsInput, output *ec2.DescribeVpcPeeringConnectionsOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, connection := range output.VpcPeeringConnections {
		attributes, err := adapterhelpers.ToAttributesWithExclude(connection, "tags")

		if err != nil {
			return nil, err
		}

		item := sdp.Item{
			Type:            "ec2-vpc-peering-connection",
			UniqueAttribute: "VpcPeeringConnectionId",
			Scope:           scope,
			Attributes:      attributes,
			Tags:            tagsToMap(connection.Tags),
		}

		if connection.Status != nil {
			switch connection.Status.Code {
			case types.VpcPeeringConnectionStateReasonCodeInitiatingRequest:
				item.Health = sdp.Health_HEALTH_PENDING.Enum()
			case types.VpcPeeringConnectionStateReasonCodePendingAcceptance:
				item.Health = sdp.Health_HEALTH_PENDING.Enum()
			case types.VpcPeeringConnectionStateReasonCodeActive:
				item.Health = sdp.Health_HEALTH_OK.Enum()
			case types.VpcPeeringConnectionStateReasonCodeDeleted:
				item.Health = sdp.Health_HEALTH_UNKNOWN.Enum()
			case types.VpcPeeringConnectionStateReasonCodeRejected:
				item.Health = sdp.Health_HEALTH_ERROR.Enum()
			case types.VpcPeeringConnectionStateReasonCodeFailed:
				item.Health = sdp.Health_HEALTH_ERROR.Enum()
			case types.VpcPeeringConnectionStateReasonCodeExpired:
				item.Health = sdp.Health_HEALTH_ERROR.Enum()
			case types.VpcPeeringConnectionStateReasonCodeProvisioning:
				item.Health = sdp.Health_HEALTH_PENDING.Enum()
			case types.VpcPeeringConnectionStateReasonCodeDeleting:
				item.Health = sdp.Health_HEALTH_WARNING.Enum()
			}
		}

		if connection.AccepterVpcInfo != nil {
			if connection.AccepterVpcInfo.Region != nil {
				if connection.AccepterVpcInfo.VpcId != nil && connection.AccepterVpcInfo.OwnerId != nil {
					pairedScope := adapterhelpers.FormatScope(*connection.AccepterVpcInfo.OwnerId, *connection.AccepterVpcInfo.Region)

					// +overmind:link ec2-vpc
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
						Query: &sdp.Query{
							Type:   "ec2-vpc",
							Method: sdp.QueryMethod_GET,
							Query:  *connection.AccepterVpcInfo.VpcId,
							Scope:  pairedScope,
						},
						BlastPropagation: &sdp.BlastPropagation{
							// The VPC will affect everything in it
							In: true,
							// We can't affect the VPC
							Out: false,
						},
					})
				}
			}

		}

		if connection.RequesterVpcInfo != nil {
			if connection.RequesterVpcInfo.Region != nil {
				if connection.RequesterVpcInfo.VpcId != nil && connection.RequesterVpcInfo.OwnerId != nil {
					pairedScope := adapterhelpers.FormatScope(*connection.RequesterVpcInfo.OwnerId, *connection.RequesterVpcInfo.Region)

					// +overmind:link ec2-vpc
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
						Query: &sdp.Query{
							Type:   "ec2-vpc",
							Method: sdp.QueryMethod_GET,
							Query:  *connection.RequesterVpcInfo.VpcId,
							Scope:  pairedScope,
						},
						BlastPropagation: &sdp.BlastPropagation{
							// The VPC will affect everything in it
							In: true,
							// We can't affect the VPC
							Out: false,
						},
					})
				}
			}
		}

		items = append(items, &item)
	}

	return items, nil
}

//go:generate docgen ../../docs-data
// +overmind:type ec2-vpc-peering-connection
// +overmind:descriptiveType VPC Peering Connection
// +overmind:get Get VPC Peering Connection by ID
// +overmind:list List VPC Peering Connections
// +overmind:search Search VPC Peering Connections by ARN
// +overmind:group AWS
// +overmind:terraform:queryMap aws_vpc_peering_connection.id
// +overmind:terraform:queryMap aws_vpc_peering_connection_accepter.id
// +overmind:terraform:queryMap aws_vpc_peering_connection_options.vpc_peering_connection_id

func NewVpcPeeringConnectionAdapter(client *ec2.Client, accountID string, region string) *adapterhelpers.DescribeOnlyAdapter[*ec2.DescribeVpcPeeringConnectionsInput, *ec2.DescribeVpcPeeringConnectionsOutput, *ec2.Client, *ec2.Options] {
	return &adapterhelpers.DescribeOnlyAdapter[*ec2.DescribeVpcPeeringConnectionsInput, *ec2.DescribeVpcPeeringConnectionsOutput, *ec2.Client, *ec2.Options]{
		Region:          region,
		Client:          client,
		AccountID:       accountID,
		ItemType:        "ec2-vpc-peering-connection",
		AdapterMetadata: VpcPeeringConnectionMetadata(),
		DescribeFunc: func(ctx context.Context, client *ec2.Client, input *ec2.DescribeVpcPeeringConnectionsInput) (*ec2.DescribeVpcPeeringConnectionsOutput, error) {
			return client.DescribeVpcPeeringConnections(ctx, input)
		},
		InputMapperGet: func(scope, query string) (*ec2.DescribeVpcPeeringConnectionsInput, error) {
			return &ec2.DescribeVpcPeeringConnectionsInput{
				VpcPeeringConnectionIds: []string{query},
			}, nil
		},
		InputMapperList: func(scope string) (*ec2.DescribeVpcPeeringConnectionsInput, error) {
			return &ec2.DescribeVpcPeeringConnectionsInput{}, nil
		},
		PaginatorBuilder: func(client *ec2.Client, params *ec2.DescribeVpcPeeringConnectionsInput) adapterhelpers.Paginator[*ec2.DescribeVpcPeeringConnectionsOutput, *ec2.Options] {
			return ec2.NewDescribeVpcPeeringConnectionsPaginator(client, params)
		},
		OutputMapper: vpcPeeringConnectionOutputMapper,
	}
}

func VpcPeeringConnectionMetadata() sdp.AdapterMetadata {
	return sdp.AdapterMetadata{
		Type:            "ec2-vpc-peering-connection",
		DescriptiveName: "VPC Peering Connection",
		SupportedQueryMethods: &sdp.AdapterSupportedQueryMethods{
			Get:             true,
			List:            true,
			Search:          true,
			GetDescription:  "Get a VPC Peering Connection by ID",
			ListDescription: "List all VPC Peering Connections",
		},
		PotentialLinks: []string{"ec2-vpc"},
		TerraformMappings: []*sdp.TerraformMapping{
			{TerraformQueryMap: "aws_vpc_peering_connection.id"},
			{TerraformQueryMap: "aws_vpc_peering_connection_accepter.id"},
			{TerraformQueryMap: "aws_vpc_peering_connection_options.vpc_peering_connection_id"},
		},
		Category: sdp.AdapterCategory_ADAPTER_CATEGORY_NETWORK,
	}
}

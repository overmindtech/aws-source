package directconnect

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/directconnect"
	"github.com/overmindtech/aws-source/adapters"
	"github.com/overmindtech/sdp-go"
)

func directConnectGatewayAssociationProposalOutputMapper(_ context.Context, _ *directconnect.Client, scope string, _ *directconnect.DescribeDirectConnectGatewayAssociationProposalsInput, output *directconnect.DescribeDirectConnectGatewayAssociationProposalsOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, associationProposal := range output.DirectConnectGatewayAssociationProposals {
		attributes, err := adapters.ToAttributesWithExclude(associationProposal, "tags")
		if err != nil {
			return nil, err
		}

		item := sdp.Item{
			Type:            "directconnect-direct-connect-gateway-association-proposal",
			UniqueAttribute: "ProposalId",
			Attributes:      attributes,
			Scope:           scope,
		}

		if associationProposal.DirectConnectGatewayId != nil && associationProposal.AssociatedGateway != nil {
			// +overmind:link directconnect-direct-connect-gateway-association
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "directconnect-direct-connect-gateway-association",
					Method: sdp.QueryMethod_GET,
					Query:  fmt.Sprintf("%s/%s", *associationProposal.DirectConnectGatewayId, *associationProposal.AssociatedGateway.Id),
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					// Any change on the association won't have an impact on the proposal
					// Its life cycle ends when the association is accepted or rejected
					In: true,
					// Accepting a proposal will establish the association
					Out: true,
				},
			})
		}

		items = append(items, &item)
	}

	return items, nil
}

//go:generate docgen ../../docs-data
// +overmind:type directconnect-direct-connect-gateway-association-proposal
// +overmind:descriptiveType Direct Connect Gateway Association Proposal
// +overmind:get Get a Direct Connect Gateway Association Proposal by ID
// +overmind:list List all Direct Connect Gateway Association Proposals
// +overmind:search Search Direct Connect Gateway Association Proposals by ARN
// +overmind:group AWS
// +overmind:terraform:queryMap aws_dx_gateway_association_proposal.id

func NewDirectConnectGatewayAssociationProposalSource(client *directconnect.Client, accountID string, region string) *adapters.DescribeOnlySource[*directconnect.DescribeDirectConnectGatewayAssociationProposalsInput, *directconnect.DescribeDirectConnectGatewayAssociationProposalsOutput, *directconnect.Client, *directconnect.Options] {
	return &adapters.DescribeOnlySource[*directconnect.DescribeDirectConnectGatewayAssociationProposalsInput, *directconnect.DescribeDirectConnectGatewayAssociationProposalsOutput, *directconnect.Client, *directconnect.Options]{
		Region:          region,
		Client:          client,
		AccountID:       accountID,
		AdapterMetadata: DirectConnectGatewayAssociationProposalMetadata(),
		ItemType:        "directconnect-direct-connect-gateway-association-proposal",
		DescribeFunc: func(ctx context.Context, client *directconnect.Client, input *directconnect.DescribeDirectConnectGatewayAssociationProposalsInput) (*directconnect.DescribeDirectConnectGatewayAssociationProposalsOutput, error) {
			return client.DescribeDirectConnectGatewayAssociationProposals(ctx, input)
		},
		InputMapperGet: func(scope, query string) (*directconnect.DescribeDirectConnectGatewayAssociationProposalsInput, error) {
			return &directconnect.DescribeDirectConnectGatewayAssociationProposalsInput{
				ProposalId: &query,
			}, nil
		},
		InputMapperList: func(scope string) (*directconnect.DescribeDirectConnectGatewayAssociationProposalsInput, error) {
			return &directconnect.DescribeDirectConnectGatewayAssociationProposalsInput{}, nil
		},
		OutputMapper: directConnectGatewayAssociationProposalOutputMapper,
	}
}

func DirectConnectGatewayAssociationProposalMetadata() sdp.AdapterMetadata {
	return sdp.AdapterMetadata{
		DescriptiveName: "Direct Connect Gateway Association Proposal",
		Type:            "directconnect-direct-connect-gateway-association-proposal",
		SupportedQueryMethods: &sdp.AdapterSupportedQueryMethods{
			Get:               true,
			List:              true,
			Search:            true,
			GetDescription:    "Get a Direct Connect Gateway Association Proposal by ID",
			ListDescription:   "List all Direct Connect Gateway Association Proposals",
			SearchDescription: "Search Direct Connect Gateway Association Proposals by ARN",
		},
		TerraformMappings: []*sdp.TerraformMapping{
			{TerraformQueryMap: "aws_dx_gateway_association_proposal.id"},
		},
		Category:       sdp.AdapterCategory_ADAPTER_CATEGORY_CONFIGURATION,
		PotentialLinks: []string{"directconnect-direct-connect-gateway-association"},
	}
}

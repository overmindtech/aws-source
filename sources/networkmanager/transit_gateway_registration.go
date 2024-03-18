package networkmanager

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func transitGatewayRegistrationOutputMapper(_ context.Context, _ *networkmanager.Client, scope string, _ *networkmanager.GetTransitGatewayRegistrationsInput, output *networkmanager.GetTransitGatewayRegistrationsOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, r := range output.TransitGatewayRegistrations {
		var err error
		var attrs *sdp.ItemAttributes
		attrs, err = sources.ToAttributesCase(r, "tags")

		if err != nil {
			return nil, &sdp.QueryError{
				ErrorType:   sdp.QueryError_OTHER,
				ErrorString: err.Error(),
				Scope:       scope,
			}
		}

		attrs.Set("globalNetworkIdWithTransitGatewayARN", idWithGlobalNetwork(*r.GlobalNetworkId, *r.TransitGatewayArn))

		item := sdp.Item{
			Type:            "networkmanager-transit-gateway-registration",
			UniqueAttribute: "globalNetworkIdWithTransitGatewayARN",
			Scope:           scope,
			Attributes:      attrs,
			LinkedItemQueries: []*sdp.LinkedItemQuery{
				{
					Query: &sdp.Query{
						// +overmind:link networkmanager-global-network
						Type:   "networkmanager-global-network",
						Method: sdp.QueryMethod_GET,
						Query:  *r.GlobalNetworkId,
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						In:  true,
						Out: false,
					},
				},
			},
		}

		// TODO: add support for ec2-transit-gateway
		// ARN example: "arn:aws:ec2:us-west-2:123456789012:transit-gateway/tgw-1234"
		if r.TransitGatewayArn != nil {
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					// +overmind:link ec2-transit-gateway
					Type:   "ec2-transit-gateway",
					Method: sdp.QueryMethod_SEARCH,
					Query:  *r.TransitGatewayArn,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					In:  true,
					Out: false,
				},
			})
		}

		items = append(items, &item)
	}

	return items, nil
}

//go:generate docgen ../../docs-data
// +overmind:type networkmanager-transit-gateway-registration
// +overmind:descriptiveType Networkmanager Transit Gateway Registrations
// +overmind:get Get a Networkmanager Transit Gateway Registrations
// +overmind:list List all Networkmanager Transit Gateway Registrations
// +overmind:search Search for Networkmanager Transit Gateway Registrations by GlobalNetworkId
// +overmind:group AWS

func NewTransitGatewayRegistrationSource(client *networkmanager.Client, accountID, region string) *sources.DescribeOnlySource[*networkmanager.GetTransitGatewayRegistrationsInput, *networkmanager.GetTransitGatewayRegistrationsOutput, *networkmanager.Client, *networkmanager.Options] {
	return &sources.DescribeOnlySource[*networkmanager.GetTransitGatewayRegistrationsInput, *networkmanager.GetTransitGatewayRegistrationsOutput, *networkmanager.Client, *networkmanager.Options]{
		Client:    client,
		AccountID: accountID,
		Region:    region,
		ItemType:  "networkmanager-transit-gateway-registration",
		DescribeFunc: func(ctx context.Context, client *networkmanager.Client, input *networkmanager.GetTransitGatewayRegistrationsInput) (*networkmanager.GetTransitGatewayRegistrationsOutput, error) {
			return client.GetTransitGatewayRegistrations(ctx, input)
		},
		InputMapperGet: func(scope, query string) (*networkmanager.GetTransitGatewayRegistrationsInput, error) {
			sections := strings.Split(query, "|")
			switch len(sections) {
			case 1:
				// only GlobalNetworkId
				return &networkmanager.GetTransitGatewayRegistrationsInput{
					GlobalNetworkId: &sections[0],
				}, nil
			case 2:
				// we are using a custom id of {globalNetworkId}|{ec2-transit-gateway.ID}
				// e.g. searching from ec2-transit-gateway
				return &networkmanager.GetTransitGatewayRegistrationsInput{
					GlobalNetworkId: &sections[0],
					TransitGatewayArns: []string{
						sections[1],
					},
				}, nil
			default:
				return nil, &sdp.QueryError{
					ErrorType:   sdp.QueryError_NOTFOUND,
					ErrorString: "invalid query for networkmanager-transit-gateway-registration get function",
				}
			}
		},
		InputMapperList: func(scope string) (*networkmanager.GetTransitGatewayRegistrationsInput, error) {
			return nil, &sdp.QueryError{
				ErrorType:   sdp.QueryError_NOTFOUND,
				ErrorString: "list not supported for networkmanager-transit-gateway-registration, use search",
			}
		},
		PaginatorBuilder: func(client *networkmanager.Client, params *networkmanager.GetTransitGatewayRegistrationsInput) sources.Paginator[*networkmanager.GetTransitGatewayRegistrationsOutput, *networkmanager.Options] {
			return networkmanager.NewGetTransitGatewayRegistrationsPaginator(client, params)
		},
		OutputMapper: transitGatewayRegistrationOutputMapper,
		InputMapperSearch: func(ctx context.Context, client *networkmanager.Client, scope, query string) (*networkmanager.GetTransitGatewayRegistrationsInput, error) {
			// Search by GlobalNetworkId
			return &networkmanager.GetTransitGatewayRegistrationsInput{
				GlobalNetworkId: &query,
			}, nil
		},
	}
}

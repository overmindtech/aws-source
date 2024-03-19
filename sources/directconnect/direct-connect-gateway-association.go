package directconnect

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/directconnect"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

const (
	directConnectGatewayIDVirtualGatewayIDFormat = "direct_connect_gateway_id/virtual_gateway_id"
	virtualGatewayIDFormat                       = "virtual_gateway_id"
)

func directConnectGatewayAssociationOutputMapper(_ context.Context, _ *directconnect.Client, scope string, _ *directconnect.DescribeDirectConnectGatewayAssociationsInput, output *directconnect.DescribeDirectConnectGatewayAssociationsOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, association := range output.DirectConnectGatewayAssociations {
		attributes, err := sources.ToAttributesCase(association, "tags")
		if err != nil {
			return nil, err
		}

		item := sdp.Item{
			Type:            "directconnect-direct-connect-gateway-association",
			UniqueAttribute: "associationId",
			Attributes:      attributes,
			Scope:           scope,
		}

		// stateChangeError =>The error message if the state of an object failed to advance.
		if association.StateChangeError != nil {
			item.Health = sdp.Health_HEALTH_ERROR.Enum()
		} else {
			item.Health = sdp.Health_HEALTH_OK.Enum()
		}

		if association.DirectConnectGatewayId != nil {
			// +overmind:link directconnect-direct-connect-gateway
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "directconnect-direct-connect-gateway",
					Method: sdp.QueryMethod_GET,
					Query:  *association.DirectConnectGatewayId,
					Scope:  "global",
				},
				BlastPropagation: &sdp.BlastPropagation{
					// Deleting a direct connect gateway will change the state of the association
					In: true,
					// We can't affect the direct connect gateway
					Out: false,
				},
			})
		}

		if association.VirtualGatewayId != nil {
			// +overmind:link directconnect-virtual-gateway
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "directconnect-virtual-gateway",
					Method: sdp.QueryMethod_GET,
					Query:  *association.VirtualGatewayId,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					// Deleting a virtual gateway will change the state of the association
					In: true,
					// We can't affect the virtual gateway
					Out: false,
				},
			})
		}

		items = append(items, &item)
	}

	return items, nil
}

//go:generate docgen ../../docs-data
// +overmind:type directconnect-direct-connect-gateway-association
// +overmind:descriptiveType Direct Connect Gateway Association
// +overmind:get Get a direct connect gateway association by direct connect gateway ID and virtual gateway ID
// +overmind:search Search direct connect gateway associations by direct connect gateway ID
// +overmind:group AWS
// +overmind:terraform:queryMap aws_dx_gateway_association.id

func NewDirectConnectGatewayAssociationSource(client *directconnect.Client, accountID string, region string, limit *sources.LimitBucket) *sources.DescribeOnlySource[*directconnect.DescribeDirectConnectGatewayAssociationsInput, *directconnect.DescribeDirectConnectGatewayAssociationsOutput, *directconnect.Client, *directconnect.Options] {
	return &sources.DescribeOnlySource[*directconnect.DescribeDirectConnectGatewayAssociationsInput, *directconnect.DescribeDirectConnectGatewayAssociationsOutput, *directconnect.Client, *directconnect.Options]{
		Region:    region,
		Client:    client,
		AccountID: accountID,
		ItemType:  "directconnect-direct-connect-gateway-association",
		DescribeFunc: func(ctx context.Context, client *directconnect.Client, input *directconnect.DescribeDirectConnectGatewayAssociationsInput) (*directconnect.DescribeDirectConnectGatewayAssociationsOutput, error) {
			limit.Wait(ctx) // Wait for rate limiting
			return client.DescribeDirectConnectGatewayAssociations(ctx, input)
		},
		InputMapperGet: func(scope, query string) (*directconnect.DescribeDirectConnectGatewayAssociationsInput, error) {
			// query must be either:
			// - in the format of "directConnectGatewayID/virtualGatewayID"
			// - virtualGatewayID => associatedGatewayID
			dxGatewayID, virtualGatewayID, err := parseDirectConnectGatewayAssociationGetInputQuery(query)
			if err != nil {
				return nil, &sdp.QueryError{
					ErrorType:   sdp.QueryError_NOTFOUND,
					ErrorString: err.Error(),
				}
			}

			if dxGatewayID != "" {
				return &directconnect.DescribeDirectConnectGatewayAssociationsInput{
					DirectConnectGatewayId: &dxGatewayID,
					VirtualGatewayId:       &virtualGatewayID,
				}, nil
			} else {
				return &directconnect.DescribeDirectConnectGatewayAssociationsInput{
					AssociatedGatewayId: &virtualGatewayID,
				}, nil
			}
		},
		InputMapperList: func(scope string) (*directconnect.DescribeDirectConnectGatewayAssociationsInput, error) {
			return nil, &sdp.QueryError{
				ErrorType:   sdp.QueryError_NOTFOUND,
				ErrorString: "list not supported for directconnect-direct-connect-gateway-association, use search",
			}
		},
		OutputMapper: directConnectGatewayAssociationOutputMapper,
		InputMapperSearch: func(ctx context.Context, client *directconnect.Client, scope, query string) (*directconnect.DescribeDirectConnectGatewayAssociationsInput, error) {
			return &directconnect.DescribeDirectConnectGatewayAssociationsInput{
				DirectConnectGatewayId: &query,
			}, nil
		},
	}
}

// parseDirectConnectGatewayAssociationGetInputQuery expects a query:
//   - in the format of "directConnectGatewayID/virtualGatewayID"
//   - virtualGatewayID => associatedGatewayID
//
// First returned item is directConnectGatewayID, second is virtualGatewayID
func parseDirectConnectGatewayAssociationGetInputQuery(query string) (string, string, error) {
	ids := strings.Split(query, "/")
	switch len(ids) {
	case 1:
		return "", ids[0], nil
	case 2:
		return ids[0], ids[1], nil
	default:
		return "", "", fmt.Errorf("invalid query, expected in the format of %s or %s, got: %s", directConnectGatewayIDVirtualGatewayIDFormat, virtualGatewayIDFormat, query)
	}
}

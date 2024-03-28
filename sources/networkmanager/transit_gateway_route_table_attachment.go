package networkmanager

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func getTransitGatewayRouteTableAttachmentGetFunc(ctx context.Context, client *networkmanager.Client, _, query string) (*types.TransitGatewayRouteTableAttachment, error) {
	out, err := client.GetTransitGatewayRouteTableAttachment(ctx, &networkmanager.GetTransitGatewayRouteTableAttachmentInput{
		AttachmentId: &query,
	})
	if err != nil {
		return nil, err
	}

	return out.TransitGatewayRouteTableAttachment, nil
}

func transitGatewayRouteTableAttachmentItemMapper(scope string, awsItem *types.TransitGatewayRouteTableAttachment) (*sdp.Item, error) {
	attributes, err := sources.ToAttributesCase(awsItem)

	if err != nil {
		return nil, err
	}

	// The uniqueAttributeValue for this is a nested value of AttachmentId:
	if awsItem != nil && awsItem.Attachment != nil {
		attributes.Set("attachmentId", *awsItem.Attachment.AttachmentId)
	}

	item := sdp.Item{
		Type:            "networkmanager-transit-gateway-route-table-attachment",
		UniqueAttribute: "attachmentId",
		Attributes:      attributes,
		Scope:           scope,
	}

	if awsItem.Attachment != nil && awsItem.Attachment.CoreNetworkId != nil {
		item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
			Query: &sdp.Query{
				// +overmind:link networkmanager-core-network
				Type:   "networkmanager-core-network",
				Method: sdp.QueryMethod_GET,
				Query:  *awsItem.Attachment.CoreNetworkId,
				Scope:  scope,
			},
			BlastPropagation: &sdp.BlastPropagation{
				In:  true,
				Out: false,
			},
		})
	}

	if awsItem.PeeringId != nil {
		item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
			Query: &sdp.Query{
				// +overmind:link networkmanager-transit-gateway-peering
				Type:   "networkmanager-transit-gateway-peering",
				Method: sdp.QueryMethod_GET,
				Query:  *awsItem.PeeringId,
				Scope:  scope,
			},
			BlastPropagation: &sdp.BlastPropagation{
				In:  true,
				Out: true,
			},
		})
	}

	// TODO: add support for ec2-transit-gateway-route-table
	// ARN example: "arn:aws:ec2:us-west-2:123456789012:transit-gateway-route-table/tgw-rtb-9876543210123456"
	if awsItem.TransitGatewayRouteTableArn != nil {
		tmp := strings.Split(*awsItem.TransitGatewayRouteTableArn, "/")
		if len(tmp) == 2 {
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					// +overmind:link ec2-transit-gateway-route-table
					Type:   "ec2-transit-gateway-route-table",
					Method: sdp.QueryMethod_SEARCH,
					Query:  tmp[1],
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					In:  true,
					Out: false,
				},
			})
		}
	}

	return &item, nil
}

//go:generate docgen ../../docs-data
// +overmind:type networkmanager-transit-gateway-route-table-attachment
// +overmind:descriptiveType Networkmanager Transit Gateway Route Table Attachment
// +overmind:get Get a Networkmanager Transit Gateway Route Table Attachment by id
// +overmind:group AWS
// +overmind:terraform:queryMap aws_networkmanager_transit_gateway_route_table_attachment.id

func NewTransitGatewayRouteTableAttachmentSource(client *networkmanager.Client, accountID, region string) *sources.GetListSource[*types.TransitGatewayRouteTableAttachment, *networkmanager.Client, *networkmanager.Options] {
	return &sources.GetListSource[*types.TransitGatewayRouteTableAttachment, *networkmanager.Client, *networkmanager.Options]{
		Client:    client,
		AccountID: accountID,
		ItemType:  "networkmanager-transit-gateway-route-table-attachment",
		GetFunc: func(ctx context.Context, client *networkmanager.Client, scope string, query string) (*types.TransitGatewayRouteTableAttachment, error) {
			return getTransitGatewayRouteTableAttachmentGetFunc(ctx, client, scope, query)
		},
		ItemMapper: transitGatewayRouteTableAttachmentItemMapper,
		ListFunc: func(ctx context.Context, client *networkmanager.Client, scope string) ([]*types.TransitGatewayRouteTableAttachment, error) {
			return nil, &sdp.QueryError{
				ErrorType:   sdp.QueryError_NOTFOUND,
				ErrorString: "list not supported for networkmanager-transit-gateway-route-table-attachment, use get",
			}
		},
	}
}

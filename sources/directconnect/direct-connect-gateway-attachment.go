package directconnect

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/directconnect"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func directConnectGatewayAttachmentOutputMapper(_ context.Context, _ *directconnect.Client, scope string, _ *directconnect.DescribeDirectConnectGatewayAttachmentsInput, output *directconnect.DescribeDirectConnectGatewayAttachmentsOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, attachment := range output.DirectConnectGatewayAttachments {
		attributes, err := sources.ToAttributesWithExclude(attachment, "tags")
		if err != nil {
			return nil, err
		}

		// The uniqueAttributeValue for this is a custom field:
		// {gatewayId}/{virtualInterfaceId}
		// i.e., "cf68415c-f4ae-48f2-87a7-3b52cexample/dxvif-ffhhk74f"
		err = attributes.Set("UniqueName", fmt.Sprintf("%s/%s", *attachment.DirectConnectGatewayId, *attachment.VirtualInterfaceId))
		if err != nil {
			return nil, err
		}

		item := sdp.Item{
			Type:            "directconnect-direct-connect-gateway-attachment",
			UniqueAttribute: "UniqueName",
			Attributes:      attributes,
			Scope:           scope,
		}

		// stateChangeError =>The error message if the state of an object failed to advance.
		if attachment.StateChangeError != nil {
			item.Health = sdp.Health_HEALTH_ERROR.Enum()
		} else {
			item.Health = sdp.Health_HEALTH_OK.Enum()
		}

		if attachment.DirectConnectGatewayId != nil {
			// +overmind:link directconnect-direct-connect-gateway
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "directconnect-direct-connect-gateway",
					Method: sdp.QueryMethod_GET,
					Query:  *attachment.DirectConnectGatewayId,
					Scope:  "global",
				},
				BlastPropagation: &sdp.BlastPropagation{
					// This is not clearly documented, but it seems that if the gateway is deleted, the attachment state will change to detaching
					In: true,
					// We can't affect the direct connect gateway
					Out: false,
				},
			})
		}

		if attachment.VirtualInterfaceId != nil {
			// +overmind:link directconnect-virtual-interface
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "directconnect-virtual-interface",
					Method: sdp.QueryMethod_GET,
					Query:  *attachment.VirtualInterfaceId,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					// If virtual interface is deleted, the attachment state will change to detaching
					// https://docs.aws.amazon.com/directconnect/latest/APIReference/API_DirectConnectGatewayAttachment.html#API_DirectConnectGatewayAttachment_Contents
					In: true,
					// We can't affect the virtual interface
					Out: false,
				},
			})
		}

		items = append(items, &item)
	}

	return items, nil
}

//go:generate docgen ../../docs-data
// +overmind:type directconnect-direct-connect-gateway-attachment
// +overmind:descriptiveType Direct Connect Gateway Attachment
// +overmind:get Get a direct connect gateway attachment by "DirectConnectGatewayId/VirtualInterfaceId"
// +overmind:search Search direct connect gateway attachments for given VirtualInterfaceId
// +overmind:group AWS

func NewDirectConnectGatewayAttachmentSource(client *directconnect.Client, accountID string, region string) *sources.DescribeOnlySource[*directconnect.DescribeDirectConnectGatewayAttachmentsInput, *directconnect.DescribeDirectConnectGatewayAttachmentsOutput, *directconnect.Client, *directconnect.Options] {
	return &sources.DescribeOnlySource[*directconnect.DescribeDirectConnectGatewayAttachmentsInput, *directconnect.DescribeDirectConnectGatewayAttachmentsOutput, *directconnect.Client, *directconnect.Options]{
		Region:    region,
		Client:    client,
		AccountID: accountID,
		ItemType:  "directconnect-direct-connect-gateway-attachment",
		DescribeFunc: func(ctx context.Context, client *directconnect.Client, input *directconnect.DescribeDirectConnectGatewayAttachmentsInput) (*directconnect.DescribeDirectConnectGatewayAttachmentsOutput, error) {
			return client.DescribeDirectConnectGatewayAttachments(ctx, input)
		},
		InputMapperGet: func(scope, query string) (*directconnect.DescribeDirectConnectGatewayAttachmentsInput, error) {
			gatewayID, virtualInterfaceID, err := parseGatewayIDVirtualInterfaceID(query)
			if err != nil {
				return nil, &sdp.QueryError{
					ErrorType:   sdp.QueryError_NOTFOUND,
					ErrorString: err.Error(),
				}
			}
			return &directconnect.DescribeDirectConnectGatewayAttachmentsInput{
				DirectConnectGatewayId: &gatewayID,
				VirtualInterfaceId:     &virtualInterfaceID,
			}, nil
		},
		InputMapperList: func(scope string) (*directconnect.DescribeDirectConnectGatewayAttachmentsInput, error) {
			return nil, &sdp.QueryError{
				ErrorType:   sdp.QueryError_NOTFOUND,
				ErrorString: "list not supported for directconnect-direct-connect-gateway-attachment, use search",
			}
		},
		OutputMapper: directConnectGatewayAttachmentOutputMapper,
		InputMapperSearch: func(ctx context.Context, client *directconnect.Client, scope, query string) (*directconnect.DescribeDirectConnectGatewayAttachmentsInput, error) {
			return &directconnect.DescribeDirectConnectGatewayAttachmentsInput{
				VirtualInterfaceId: &query,
			}, nil
		},
	}
}

// parseGatewayIDVirtualInterfaceID expects a query in the format of "gatewayID/virtualInterfaceID"
// First returned item is gatewayID, second is virtualInterfaceID
func parseGatewayIDVirtualInterfaceID(query string) (string, string, error) {
	ids := strings.Split(query, "/")
	if len(ids) != 2 {
		return "", "", fmt.Errorf("invalid query, expected in the format of %s, got: %s", gatewayIDVirtualInterfaceIDFormat, query)
	}

	return ids[0], ids[1], nil
}

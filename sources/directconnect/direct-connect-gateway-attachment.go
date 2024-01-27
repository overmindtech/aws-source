package directconnect

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/directconnect"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func directConnectGatewayAttachmentOutputMapper(_ context.Context, _ *directconnect.Client, scope string, _ *directconnect.DescribeDirectConnectGatewayAttachmentsInput, output *directconnect.DescribeDirectConnectGatewayAttachmentsOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, attachment := range output.DirectConnectGatewayAttachments {
		attributes, err := sources.ToAttributesCase(attachment, "tags")
		if err != nil {
			return nil, err
		}

		// The uniqueAttributeValue for this is a custom field:
		// {gatewayId} {virtualInterfaceId}
		// i.e., "cf68415c-f4ae-48f2-87a7-3b52cexample dxvif-ffhhk74f"
		err = attributes.Set("uniqueName", fmt.Sprintf(gatewayIDVirtualInterfaceIDFmt, *attachment.DirectConnectGatewayId, *attachment.VirtualInterfaceId))
		if err != nil {
			return nil, err
		}

		item := sdp.Item{
			Type:            "directconnect-direct-connect-gateway-attachment",
			UniqueAttribute: "uniqueName",
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
// +overmind:get Get a direct connect gateway attachment by DirectConnectGatewayId and VirtualInterfaceId
// +overmind:search Search direct connect gateway attachments for given VirtualInterfaceId
// +overmind:group AWS

func NewDirectConnectGatewayAttachmentSource(config aws.Config, accountID string, limit *sources.LimitBucket) *sources.DescribeOnlySource[*directconnect.DescribeDirectConnectGatewayAttachmentsInput, *directconnect.DescribeDirectConnectGatewayAttachmentsOutput, *directconnect.Client, *directconnect.Options] {
	return &sources.DescribeOnlySource[*directconnect.DescribeDirectConnectGatewayAttachmentsInput, *directconnect.DescribeDirectConnectGatewayAttachmentsOutput, *directconnect.Client, *directconnect.Options]{
		Config:    config,
		Client:    directconnect.NewFromConfig(config),
		AccountID: accountID,
		ItemType:  "directconnect-direct-connect-gateway-attachment",
		DescribeFunc: func(ctx context.Context, client *directconnect.Client, input *directconnect.DescribeDirectConnectGatewayAttachmentsInput) (*directconnect.DescribeDirectConnectGatewayAttachmentsOutput, error) {
			limit.Wait(ctx) // Wait for rate limiting
			return client.DescribeDirectConnectGatewayAttachments(ctx, input)
		},
		InputMapperGet: func(scope, query string) (*directconnect.DescribeDirectConnectGatewayAttachmentsInput, error) {
			gatewayID, virtualInterfaceID, err := parseGatewayIDVirtualInterfaceID(query)
			if err != nil {
				return nil, fmt.Errorf(`invalid query, expected in the format of "`+gatewayIDVirtualInterfaceIDFmt, "<some_id>", "<some_id>"+`", %w`, err)
			}
			return &directconnect.DescribeDirectConnectGatewayAttachmentsInput{
				DirectConnectGatewayId: &gatewayID,
				VirtualInterfaceId:     &virtualInterfaceID,
			}, nil
		},
		OutputMapper: directConnectGatewayAttachmentOutputMapper,
		InputMapperSearch: func(ctx context.Context, client *directconnect.Client, scope, query string) (*directconnect.DescribeDirectConnectGatewayAttachmentsInput, error) {
			return &directconnect.DescribeDirectConnectGatewayAttachmentsInput{
				VirtualInterfaceId: &query,
			}, nil
		},
	}
}

func parseGatewayIDVirtualInterfaceID(query string) (gatewayID, virtualInterfaceID string, err error) {
	_, err = fmt.Sscanf(query, gatewayIDVirtualInterfaceIDFmt, &gatewayID, &virtualInterfaceID)
	return
}

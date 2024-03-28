package networkmanager

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func connectAttachmentGetFunc(ctx context.Context, client *networkmanager.Client, _, query string) (*types.ConnectAttachment, error) {
	out, err := client.GetConnectAttachment(ctx, &networkmanager.GetConnectAttachmentInput{
		AttachmentId: &query,
	})
	if err != nil {
		return nil, err
	}

	return out.ConnectAttachment, nil
}

func connectAttachmentItemMapper(scope string, ca *types.ConnectAttachment) (*sdp.Item, error) {
	attributes, err := sources.ToAttributesCase(ca)

	if err != nil {
		return nil, err
	}

	// The uniqueAttributeValue for this is a nested value of AttachmentId:
	if ca != nil && ca.Attachment != nil {
		attributes.Set("attachmentId", *ca.Attachment.AttachmentId)
	}

	item := sdp.Item{
		Type:            "networkmanager-connect-attachment",
		UniqueAttribute: "attachmentId",
		Attributes:      attributes,
		Scope:           scope,
	}

	if ca.Attachment != nil && ca.Attachment.CoreNetworkId != nil {
		item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
			Query: &sdp.Query{
				// +overmind:link networkmanager-core-network
				Type:   "networkmanager-core-network",
				Method: sdp.QueryMethod_GET,
				Query:  *ca.Attachment.CoreNetworkId,
				Scope:  scope,
			},
			BlastPropagation: &sdp.BlastPropagation{
				In:  true,
				Out: false,
			},
		})
	}

	return &item, nil
}

//go:generate docgen ../../docs-data
// +overmind:type networkmanager-connect-attachment
// +overmind:descriptiveType Networkmanager Connect Attachment
// +overmind:get Get a Networkmanager Connect Attachment by id
// +overmind:group AWS
// +overmind:terraform:queryMap aws_networkmanager_core_network.id

func NewConnectAttachmentSource(client *networkmanager.Client, accountID, region string) *sources.GetListSource[*types.ConnectAttachment, *networkmanager.Client, *networkmanager.Options] {
	return &sources.GetListSource[*types.ConnectAttachment, *networkmanager.Client, *networkmanager.Options]{
		Client:    client,
		AccountID: accountID,
		Region:    region,
		ItemType:  "networkmanager-connect-attachment",
		GetFunc: func(ctx context.Context, client *networkmanager.Client, scope string, query string) (*types.ConnectAttachment, error) {
			return connectAttachmentGetFunc(ctx, client, scope, query)
		},
		ItemMapper: connectAttachmentItemMapper,
		ListFunc: func(ctx context.Context, client *networkmanager.Client, scope string) ([]*types.ConnectAttachment, error) {
			return nil, &sdp.QueryError{
				ErrorType:   sdp.QueryError_NOTFOUND,
				ErrorString: "list not supported for networkmanager-connect-attachment, use get",
			}
		},
	}
}

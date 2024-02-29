package networkmanager

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func vpcAttachmentGetFunc(ctx context.Context, client *networkmanager.Client, _, query string) (*types.VpcAttachment, error) {
	out, err := client.GetVpcAttachment(ctx, &networkmanager.GetVpcAttachmentInput{
		AttachmentId: &query,
	})
	if err != nil {
		return nil, err
	}

	return out.VpcAttachment, nil
}

func vpcAttachmentItemMapper(scope string, awsItem *types.VpcAttachment) (*sdp.Item, error) {
	attributes, err := sources.ToAttributesCase(awsItem)

	if err != nil {
		return nil, err
	}

	// The uniqueAttributeValue for this is a nested value of AttachmentId:
	if awsItem != nil && awsItem.Attachment != nil {
		attributes.Set("attachmentId", *awsItem.Attachment.AttachmentId)
	}

	item := sdp.Item{
		Type:            "networkmanager-vpc-attachment",
		UniqueAttribute: "attachmentId",
		Attributes:      attributes,
		Scope:           scope,
	}

	if awsItem.Attachment.CoreNetworkId != nil {
		item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
			Query: &sdp.Query{
				// +overmind:link networkmanager-core-network
				// Search for all vpc attachments with this core network
				Type:   "networkmanager-core-network",
				Method: sdp.QueryMethod_GET,
				Query:  *awsItem.Attachment.CoreNetworkId,
				Scope:  scope,
			},
			BlastPropagation: &sdp.BlastPropagation{
				// ?? vpc attachment can affect the global network (depends on meaning of "affect" in this case)
				In: true,
				// The core network will definitely affect the vpc attachment
				Out: true,
			},
		})

	}

	return &item, nil
}

//go:generate docgen ../../docs-data
// +overmind:type networkmanager-vpc-attachment
// +overmind:descriptiveType Networkmanager VPC Attachment
// +overmind:get Get a Networkmanager VPC Attachment by id
// +overmind:group AWS
// +overmind:terraform:queryMap aws_networkmanager_vpc_attachment.id

func NewVPCAttachmentSource(config aws.Config, accountID string, limit *sources.LimitBucket) *sources.GetListSource[*types.VpcAttachment, *networkmanager.Client, *networkmanager.Options] {
	return &sources.GetListSource[*types.VpcAttachment, *networkmanager.Client, *networkmanager.Options]{
		Client:    networkmanager.NewFromConfig(config),
		AccountID: accountID,
		ItemType:  "networkmanager-vpc-attachment",
		GetFunc: func(ctx context.Context, client *networkmanager.Client, scope string, query string) (*types.VpcAttachment, error) {
			limit.Wait(ctx) // Wait for rate limiting
			return vpcAttachmentGetFunc(ctx, client, scope, query)
		},
		ItemMapper: vpcAttachmentItemMapper,
		ListFunc: func(ctx context.Context, client *networkmanager.Client, scope string) ([]*types.VpcAttachment, error) {
			return nil, &sdp.QueryError{
				ErrorType:   sdp.QueryError_NOTFOUND,
				ErrorString: "list not supported for  networkmanager-vpc-attachment, use get",
			}
		},
	}
}

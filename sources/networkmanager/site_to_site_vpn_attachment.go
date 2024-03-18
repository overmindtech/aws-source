package networkmanager

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func getSiteToSiteVpnAttachmentGetFunc(ctx context.Context, client *networkmanager.Client, _, query string) (*types.SiteToSiteVpnAttachment, error) {
	out, err := client.GetSiteToSiteVpnAttachment(ctx, &networkmanager.GetSiteToSiteVpnAttachmentInput{
		AttachmentId: &query,
	})
	if err != nil {
		return nil, err
	}

	return out.SiteToSiteVpnAttachment, nil
}

func siteToSiteVpnAttachmentItemMapper(scope string, awsItem *types.SiteToSiteVpnAttachment) (*sdp.Item, error) {
	attributes, err := sources.ToAttributesCase(awsItem)

	if err != nil {
		return nil, err
	}

	// The uniqueAttributeValue for this is a nested value of peeringId:
	if awsItem != nil && awsItem.Attachment != nil {
		attributes.Set("attachmentId", *awsItem.Attachment.AttachmentId)
	}

	item := sdp.Item{
		Type:            "networkmanager-site-to-site-vpn-attachment",
		UniqueAttribute: "attachmentId",
		Attributes:      attributes,
		Scope:           scope,
	}

	if awsItem.Attachment != nil {
		if awsItem.Attachment.CoreNetworkId != nil {
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					// +overmind:link networkmanager-core-network
					// Search for core network
					Type:   "networkmanager-core-network",
					Method: sdp.QueryMethod_GET,
					Query:  *awsItem.Attachment.CoreNetworkId,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					In:  true,
					Out: true,
				},
			})
		}

		switch awsItem.Attachment.State {
		case types.AttachmentStateCreating:
			item.Health = sdp.Health_HEALTH_PENDING.Enum()
		case types.AttachmentStateAvailable:
			item.Health = sdp.Health_HEALTH_OK.Enum()
		case types.AttachmentStateDeleting:
			item.Health = sdp.Health_HEALTH_PENDING.Enum()
		case types.AttachmentStateFailed:
			item.Health = sdp.Health_HEALTH_ERROR.Enum()
		}
	}
	// TODO: add support for ec2-vpn-connection
	if awsItem.VpnConnectionArn != nil {
		item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
			Query: &sdp.Query{
				// +overmind:link ec2-vpn-connection
				Type:   "ec2-vpn-connection",
				Method: sdp.QueryMethod_SEARCH,
				Query:  *awsItem.VpnConnectionArn,
				Scope:  scope,
			},
			BlastPropagation: &sdp.BlastPropagation{
				In:  true,
				Out: true,
			},
		})
	}

	return &item, nil
}

//go:generate docgen ../../docs-data
// +overmind:type networkmanager-site-to-site-vpn-attachment
// +overmind:descriptiveType Networkmanager Site To Site Vpn Attachment
// +overmind:get Get a Networkmanager Site To Site Vpn Attachment by id
// +overmind:group AWS
// +overmind:terraform:queryMap aws_networkmanager_site_to_site_vpn_attachment.id

func NewSiteToSiteVpnAttachmentSource(client *networkmanager.Client, accountID, region string) *sources.GetListSource[*types.SiteToSiteVpnAttachment, *networkmanager.Client, *networkmanager.Options] {
	return &sources.GetListSource[*types.SiteToSiteVpnAttachment, *networkmanager.Client, *networkmanager.Options]{
		Client:    client,
		AccountID: accountID,
		Region:    region,
		ItemType:  "networkmanager-site-to-site-vpn-attachment",
		GetFunc: func(ctx context.Context, client *networkmanager.Client, scope string, query string) (*types.SiteToSiteVpnAttachment, error) {
			return getSiteToSiteVpnAttachmentGetFunc(ctx, client, scope, query)
		},
		ItemMapper: siteToSiteVpnAttachmentItemMapper,
		ListFunc: func(ctx context.Context, client *networkmanager.Client, scope string) ([]*types.SiteToSiteVpnAttachment, error) {
			return nil, &sdp.QueryError{
				ErrorType:   sdp.QueryError_NOTFOUND,
				ErrorString: "list not supported for networkmanager-site-to-site-vpn-attachment, use get",
			}
		},
	}
}

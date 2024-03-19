package cloudfront

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func originAccessControlListFunc(ctx context.Context, client *cloudfront.Client, scope string) ([]*types.OriginAccessControl, error) {
	out, err := client.ListOriginAccessControls(ctx, &cloudfront.ListOriginAccessControlsInput{})

	if err != nil {
		return nil, err
	}

	originAccessControls := make([]*types.OriginAccessControl, len(out.OriginAccessControlList.Items))

	for i, item := range out.OriginAccessControlList.Items {
		// Annoyingly the "summary" types has exactly the same information as
		// the type returned by get, but in a slightly different format. So we
		// map it to the get format here
		originAccessControls[i] = &types.OriginAccessControl{
			Id: item.Id,
			OriginAccessControlConfig: &types.OriginAccessControlConfig{
				Name:                          item.Name,
				OriginAccessControlOriginType: item.OriginAccessControlOriginType,
				SigningBehavior:               item.SigningBehavior,
				SigningProtocol:               item.SigningProtocol,
				Description:                   item.Description,
			},
		}
	}

	return originAccessControls, nil
}

func originAccessControlItemMapper(scope string, awsItem *types.OriginAccessControl) (*sdp.Item, error) {
	attributes, err := sources.ToAttributesCase(awsItem)

	if err != nil {
		return nil, err
	}

	item := sdp.Item{
		Type:            "cloudfront-origin-access-control",
		UniqueAttribute: "id",
		Attributes:      attributes,
		Scope:           scope,
	}

	return &item, nil
}

//go:generate docgen ../../docs-data
// +overmind:type cloudfront-origin-access-control
// +overmind:descriptiveType Cloudfront Origin Access Control
// +overmind:get Get Origin Access Control by ID
// +overmind:list List Origin Access Controls
// +overmind:search Origin Access Control by ARN
// +overmind:group AWS
// +overmind:terraform:queryMap aws_cloudfront_origin_access_control.id

func NewOriginAccessControlSource(client *cloudfront.Client, accountID string) *sources.GetListSource[*types.OriginAccessControl, *cloudfront.Client, *cloudfront.Options] {
	return &sources.GetListSource[*types.OriginAccessControl, *cloudfront.Client, *cloudfront.Options]{
		ItemType:  "cloudfront-origin-access-control",
		Client:    client,
		AccountID: accountID,
		Region:    "", // Cloudfront resources aren't tied to a region
		GetFunc: func(ctx context.Context, client *cloudfront.Client, scope, query string) (*types.OriginAccessControl, error) {
			out, err := client.GetOriginAccessControl(ctx, &cloudfront.GetOriginAccessControlInput{
				Id: &query,
			})

			if err != nil {
				return nil, err
			}

			return out.OriginAccessControl, nil
		},
		ListFunc:   originAccessControlListFunc,
		ItemMapper: originAccessControlItemMapper,
	}
}

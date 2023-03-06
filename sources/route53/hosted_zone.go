package route53

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func HostedZoneGetFunc(ctx context.Context, client *route53.Client, scope, query string) (*types.HostedZone, error) {
	out, err := client.GetHostedZone(ctx, &route53.GetHostedZoneInput{
		Id: &query,
	})

	if err != nil {
		return nil, err
	}

	return out.HostedZone, nil
}

func HostedZoneListFunc(ctx context.Context, client *route53.Client, scope string) ([]*types.HostedZone, error) {
	out, err := client.ListHostedZones(ctx, &route53.ListHostedZonesInput{})

	if err != nil {
		return nil, err
	}

	zones := make([]*types.HostedZone, len(out.HostedZones))

	for i, zone := range out.HostedZones {
		zones[i] = &zone
	}

	return zones, nil
}

func HostedZoneItemMapper(scope string, awsItem *types.HostedZone) (*sdp.Item, error) {
	attributes, err := sources.ToAttributesCase(awsItem)

	if err != nil {
		return nil, err
	}

	item := sdp.Item{
		Type:            "route53-hosted-zone",
		UniqueAttribute: "id",
		Attributes:      attributes,
		Scope:           scope,
		LinkedItemQueries: []*sdp.Query{
			{
				Type:   "route53-resource-record-set",
				Method: sdp.RequestMethod_SEARCH,
				Query:  *awsItem.Id,
				Scope:  scope,
			},
		},
	}

	return &item, nil
}

func NewHostedZoneSource(config aws.Config, accountID string, region string) *sources.GetListSource[*types.HostedZone, *route53.Client, *route53.Options] {
	return &sources.GetListSource[*types.HostedZone, *route53.Client, *route53.Options]{
		ItemType:   "route53-hosted-zone",
		Client:     route53.NewFromConfig(config),
		AccountID:  accountID,
		Region:     region,
		GetFunc:    HostedZoneGetFunc,
		ListFunc:   HostedZoneListFunc,
		ItemMapper: HostedZoneItemMapper,
	}
}

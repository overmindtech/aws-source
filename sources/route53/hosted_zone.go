package route53

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func hostedZoneGetFunc(ctx context.Context, client *route53.Client, scope, query string) (*types.HostedZone, error) {
	out, err := client.GetHostedZone(ctx, &route53.GetHostedZoneInput{
		Id: &query,
	})

	if err != nil {
		return nil, err
	}

	return out.HostedZone, nil
}

func hostedZoneListFunc(ctx context.Context, client *route53.Client, scope string) ([]*types.HostedZone, error) {
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

func hostedZoneItemMapper(scope string, awsItem *types.HostedZone) (*sdp.Item, error) {
	attributes, err := sources.ToAttributesCase(awsItem)

	if err != nil {
		return nil, err
	}

	item := sdp.Item{
		Type:            "route53-hosted-zone",
		UniqueAttribute: "id",
		Attributes:      attributes,
		Scope:           scope,
		LinkedItemQueries: []*sdp.LinkedItemQuery{
			{
				Query: &sdp.Query{
					// +overmind:link route53-resource-record-set
					Type:   "route53-resource-record-set",
					Method: sdp.QueryMethod_SEARCH,
					Query:  *awsItem.Id,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					// Changing the hosted zone can affect the resource record set
					In: true,
					// The resource record set won't affect the hosted zone
					Out: false,
				},
			},
		},
	}

	return &item, nil
}

//go:generate docgen ../../docs-data
// +overmind:type route53-hosted-zone
// +overmind:descriptiveType Route53 Hosted Zone
// +overmind:get Get a hosted zone by ID
// +overmind:list List all hosted zones
// +overmind:search Search for a hosted zone by ARN
// +overmind:group AWS
// +overmind:terraform:queryMap aws_route53_hosted_zone_dnssec.id
// +overmind:terraform:queryMap aws_route53_zone.zone_id
// +overmind:terraform:queryMap aws_route53_zone_association.zone_id

func NewHostedZoneSource(config aws.Config, accountID string, region string) *sources.GetListSource[*types.HostedZone, *route53.Client, *route53.Options] {
	return &sources.GetListSource[*types.HostedZone, *route53.Client, *route53.Options]{
		ItemType:   "route53-hosted-zone",
		Client:     route53.NewFromConfig(config),
		AccountID:  accountID,
		Region:     region,
		GetFunc:    hostedZoneGetFunc,
		ListFunc:   hostedZoneListFunc,
		ItemMapper: hostedZoneItemMapper,
	}
}

package route53

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func resourceRecordSetGetFunc(ctx context.Context, client *route53.Client, scope, query string) (*types.ResourceRecordSet, error) {
	return nil, errors.New("get is not supported for route53-resource-record-set. Use search")
}

// ResourceRecordSetSearchFunc Search func that accepts a hosted zone ID as a
// query
func resourceRecordSetSearchFunc(ctx context.Context, client *route53.Client, scope, query string) ([]*types.ResourceRecordSet, error) {
	out, err := client.ListResourceRecordSets(ctx, &route53.ListResourceRecordSetsInput{
		HostedZoneId: &query,
	})

	if err != nil {
		return nil, err
	}

	zones := make([]*types.ResourceRecordSet, len(out.ResourceRecordSets))

	for i, zone := range out.ResourceRecordSets {
		zones[i] = &zone
	}

	return zones, nil
}

func resourceRecordSetItemMapper(scope string, awsItem *types.ResourceRecordSet) (*sdp.Item, error) {
	attributes, err := sources.ToAttributesCase(awsItem)

	if err != nil {
		return nil, err
	}

	item := sdp.Item{
		Type:            "route53-resource-record-set",
		UniqueAttribute: "name",
		Attributes:      attributes,
		Scope:           scope,
	}

	if awsItem.AliasTarget != nil {
		if awsItem.AliasTarget.DNSName != nil {
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
				Type:   "dns",
				Method: sdp.QueryMethod_SEARCH,
				Query:  *awsItem.AliasTarget.DNSName,
				Scope:  "global",
			})
		}
	}

	return &item, nil
}

func NewResourceRecordSetSource(config aws.Config, accountID string, region string) *sources.GetListSource[*types.ResourceRecordSet, *route53.Client, *route53.Options] {
	return &sources.GetListSource[*types.ResourceRecordSet, *route53.Client, *route53.Options]{
		ItemType:    "route53-resource-record-set",
		Client:      route53.NewFromConfig(config),
		DisableList: true,
		AccountID:   accountID,
		Region:      region,
		GetFunc:     resourceRecordSetGetFunc,
		ItemMapper:  resourceRecordSetItemMapper,
		SearchFunc:  resourceRecordSetSearchFunc,
	}
}

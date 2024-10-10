package route53

import (
	"context"
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/overmindtech/aws-source/adapters"
	"github.com/overmindtech/sdp-go"
)

func resourceRecordSetGetFunc(ctx context.Context, client *route53.Client, scope, query string) (*types.ResourceRecordSet, error) {
	return nil, errors.New("get is not supported for route53-resource-record-set. Use search")
}

// ResourceRecordSetSearchFunc Search func that accepts a hosted zone or a
// terraform ID in the format {hostedZone}_{recordName}_{type}
func resourceRecordSetSearchFunc(ctx context.Context, client *route53.Client, scope, query string) ([]*types.ResourceRecordSet, error) {
	splits := strings.Split(query, "_")

	var out *route53.ListResourceRecordSetsOutput
	var err error
	if len(splits) == 3 {
		// In this case we have a terraform ID
		var max int32 = 1
		out, err = client.ListResourceRecordSets(ctx, &route53.ListResourceRecordSetsInput{
			HostedZoneId:    &splits[0],
			StartRecordName: &splits[1],
			StartRecordType: types.RRType(splits[2]),
			MaxItems:        &max,
		})
	} else {
		// In this case we have a hosted zone ID
		out, err = client.ListResourceRecordSets(ctx, &route53.ListResourceRecordSetsInput{
			HostedZoneId: &query,
		})
	}

	if err != nil {
		return nil, err
	}

	records := make([]*types.ResourceRecordSet, 0, len(out.ResourceRecordSets))

	for _, record := range out.ResourceRecordSets {
		records = append(records, &record)
	}

	return records, nil
}

func resourceRecordSetItemMapper(_, scope string, awsItem *types.ResourceRecordSet) (*sdp.Item, error) {
	attributes, err := adapters.ToAttributesWithExclude(awsItem)

	if err != nil {
		return nil, err
	}

	item := sdp.Item{
		Type:            "route53-resource-record-set",
		UniqueAttribute: "Name",
		Attributes:      attributes,
		Scope:           scope,
	}

	if awsItem.AliasTarget != nil {
		if awsItem.AliasTarget.DNSName != nil {
			// +overmind:link dns
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "dns",
					Method: sdp.QueryMethod_SEARCH,
					Query:  *awsItem.AliasTarget.DNSName,
					Scope:  "global",
				},
				BlastPropagation: &sdp.BlastPropagation{
					// DNS aliases links
					In:  true,
					Out: true,
				},
			})
		}
	}

	for _, record := range awsItem.ResourceRecords {
		if record.Value != nil {
			// +overmind:link dns
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "dns",
					Method: sdp.QueryMethod_SEARCH,
					Query:  *record.Value,
					Scope:  "global",
				},
				BlastPropagation: &sdp.BlastPropagation{
					// DNS aliases links
					In:  true,
					Out: true,
				},
			})
		}
	}

	if awsItem.HealthCheckId != nil {
		// +overmind:link route53-health-check
		item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
			Query: &sdp.Query{
				Type:   "route53-health-check",
				Method: sdp.QueryMethod_GET,
				Query:  *awsItem.HealthCheckId,
				Scope:  scope,
			},
			BlastPropagation: &sdp.BlastPropagation{
				// Health check links tightly
				In:  true,
				Out: true,
			},
		})
	}

	return &item, nil
}

//go:generate docgen ../../docs-data
// +overmind:type route53-resource-record-set
// +overmind:descriptiveType Route53 Record Set
// +overmind:get Get a Route53 record Set by name
// +overmind:list List all record sets
// +overmind:search Search for a record set by hosted zone ID in the format "/hostedzone/JJN928734JH7HV" or "JJN928734JH7HV" or by terraform ID in the format "{hostedZone}_{recordName}_{type}"
// +overmind:group AWS
// +overmind:terraform:queryMap aws_route53_record.arn
// +overmind:terraform:queryMap aws_route53_record.id
// +overmind:terraform:method SEARCH

func NewResourceRecordSetSource(client *route53.Client, accountID string, region string) *adapters.GetListSource[*types.ResourceRecordSet, *route53.Client, *route53.Options] {
	return &adapters.GetListSource[*types.ResourceRecordSet, *route53.Client, *route53.Options]{
		ItemType:        "route53-resource-record-set",
		Client:          client,
		DisableList:     true,
		AccountID:       accountID,
		Region:          region,
		GetFunc:         resourceRecordSetGetFunc,
		ItemMapper:      resourceRecordSetItemMapper,
		SearchFunc:      resourceRecordSetSearchFunc,
		AdapterMetadata: ResourceRecordSetMetadata(),
	}
}

func ResourceRecordSetMetadata() sdp.AdapterMetadata {
	return sdp.AdapterMetadata{
		Type:            "route53-resource-record-set",
		DescriptiveName: "Route53 Record Set",
		SupportedQueryMethods: &sdp.AdapterSupportedQueryMethods{
			Get:               true,
			List:              true,
			Search:            true,
			GetDescription:    "Get a Route53 record Set by name",
			ListDescription:   "List all record sets",
			SearchDescription: "Search for a record set by hosted zone ID in the format \"/hostedzone/JJN928734JH7HV\" or \"JJN928734JH7HV\" or by terraform ID in the format \"{hostedZone}_{recordName}_{type}\"",
		},
		Category:       sdp.AdapterCategory_ADAPTER_CATEGORY_NETWORK,
		PotentialLinks: []string{"dns", "route53-health-check"},
		TerraformMappings: []*sdp.TerraformMapping{
			{TerraformQueryMap: "aws_route53_record.arn", TerraformMethod: sdp.QueryMethod_SEARCH},
			{TerraformQueryMap: "aws_route53_record.id", TerraformMethod: sdp.QueryMethod_SEARCH},
		},
	}
}

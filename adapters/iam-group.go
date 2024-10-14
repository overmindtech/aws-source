package adapters

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"

	"github.com/overmindtech/aws-source/adapterhelpers"
	"github.com/overmindtech/sdp-go"
)

func groupGetFunc(ctx context.Context, client *iam.Client, _, query string) (*types.Group, error) {
	out, err := client.GetGroup(ctx, &iam.GetGroupInput{
		GroupName: &query,
	})

	if err != nil {
		return nil, err
	}

	return out.Group, nil
}

func groupListFunc(ctx context.Context, client *iam.Client, _ string) ([]*types.Group, error) {
	out, err := client.ListGroups(ctx, &iam.ListGroupsInput{})

	if err != nil {
		return nil, err
	}

	zones := make([]*types.Group, 0, len(out.Groups))

	for i := range out.Groups {
		zones = append(zones, &out.Groups[i])
	}

	return zones, nil
}

func groupItemMapper(_, scope string, awsItem *types.Group) (*sdp.Item, error) {
	attributes, err := adapterhelpers.ToAttributesWithExclude(awsItem)

	if err != nil {
		return nil, err
	}

	item := sdp.Item{
		Type:            "iam-group",
		UniqueAttribute: "GroupName",
		Attributes:      attributes,
		Scope:           scope,
	}

	return &item, nil
}

func NewIAMGroupAdapter(client *iam.Client, accountID string, region string) *adapterhelpers.GetListAdapter[*types.Group, *iam.Client, *iam.Options] {
	return &adapterhelpers.GetListAdapter[*types.Group, *iam.Client, *iam.Options]{
		ItemType:        "iam-group",
		Client:          client,
		CacheDuration:   3 * time.Hour, // IAM has very low rate limits, we need to cache for a long time
		AccountID:       accountID,
		AdapterMetadata: iamGroupAdapterMetadata,
		GetFunc: func(ctx context.Context, client *iam.Client, scope, query string) (*types.Group, error) {
			return groupGetFunc(ctx, client, scope, query)
		},
		ListFunc: func(ctx context.Context, client *iam.Client, scope string) ([]*types.Group, error) {
			return groupListFunc(ctx, client, scope)
		},
		ItemMapper: groupItemMapper,
	}
}

var iamGroupAdapterMetadata = Metadata.Register(&sdp.AdapterMetadata{
	Type:            "iam-group",
	DescriptiveName: "IAM Group",
	SupportedQueryMethods: &sdp.AdapterSupportedQueryMethods{
		Get:               true,
		List:              true,
		Search:            true,
		GetDescription:    "Get a group by name",
		ListDescription:   "List all IAM groups",
		SearchDescription: "Search for a group by ARN",
	},
	TerraformMappings: []*sdp.TerraformMapping{
		{
			TerraformMethod:   sdp.QueryMethod_SEARCH,
			TerraformQueryMap: "aws_iam_group.arn",
		},
	},
	Category: sdp.AdapterCategory_ADAPTER_CATEGORY_SECURITY,
})

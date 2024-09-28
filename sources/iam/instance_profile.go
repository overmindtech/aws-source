package iam

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"

	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func instanceProfileGetFunc(ctx context.Context, client *iam.Client, _, query string) (*types.InstanceProfile, error) {
	out, err := client.GetInstanceProfile(ctx, &iam.GetInstanceProfileInput{
		InstanceProfileName: &query,
	})

	if err != nil {
		return nil, err
	}

	return out.InstanceProfile, nil
}

func instanceProfileListFunc(ctx context.Context, client *iam.Client, _ string) ([]*types.InstanceProfile, error) {
	out, err := client.ListInstanceProfiles(ctx, &iam.ListInstanceProfilesInput{})

	if err != nil {
		return nil, err
	}

	zones := make([]*types.InstanceProfile, 0, len(out.InstanceProfiles))

	for i := range out.InstanceProfiles {
		zones = append(zones, &out.InstanceProfiles[i])
	}

	return zones, nil
}

func instanceProfileItemMapper(_, scope string, awsItem *types.InstanceProfile) (*sdp.Item, error) {
	attributes, err := sources.ToAttributesWithExclude(awsItem)

	if err != nil {
		return nil, err
	}

	item := sdp.Item{
		Type:            "iam-instance-profile",
		UniqueAttribute: "InstanceProfileName",
		Attributes:      attributes,
		Scope:           scope,
	}

	for _, role := range awsItem.Roles {
		if arn, err := sources.ParseARN(*role.Arn); err == nil {
			// +overmind:link iam-role
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "iam-role",
					Method: sdp.QueryMethod_SEARCH,
					Query:  *role.Arn,
					Scope:  sources.FormatScope(arn.AccountID, arn.Region),
				},
				BlastPropagation: &sdp.BlastPropagation{
					// Changes to the role will affect this
					In: true,
					// We can't affect the role
					Out: false,
				},
			})
		}

		if role.PermissionsBoundary != nil {
			if arn, err := sources.ParseARN(*role.PermissionsBoundary.PermissionsBoundaryArn); err == nil {
				// +overmind:link iam-policy
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "iam-policy",
						Method: sdp.QueryMethod_SEARCH,
						Query:  *role.PermissionsBoundary.PermissionsBoundaryArn,
						Scope:  sources.FormatScope(arn.AccountID, arn.Region),
					},
					BlastPropagation: &sdp.BlastPropagation{
						// Changes to the policy will affect this
						In: true,
						// We can't affect the policy
						Out: false,
					},
				})
			}
		}
	}

	return &item, nil
}

func instanceProfileListTagsFunc(ctx context.Context, ip *types.InstanceProfile, client *iam.Client) map[string]string {
	tags := make(map[string]string)

	paginator := iam.NewListInstanceProfileTagsPaginator(client, &iam.ListInstanceProfileTagsInput{
		InstanceProfileName: ip.InstanceProfileName,
	})

	for paginator.HasMorePages() {
		out, err := paginator.NextPage(ctx)

		if err != nil {
			return sources.HandleTagsError(ctx, err)
		}

		for _, tag := range out.Tags {
			if tag.Key != nil && tag.Value != nil {
				tags[*tag.Key] = *tag.Value
			}
		}
	}

	return tags
}

//go:generate docgen ../../docs-data
// +overmind:type iam-instance-profile
// +overmind:descriptiveType IAM instance profile
// +overmind:get Get an IAM instance profile
// +overmind:list List IAM instance profiles
// +overmind:search Search IAM instance profiles by ARN
// +overmind:group AWS
// +overmind:terraform:queryMap aws_iam_instance_profile.arn
// +overmind:terraform:method SEARCH

func NewInstanceProfileSource(client *iam.Client, accountID string, region string) *sources.GetListSource[*types.InstanceProfile, *iam.Client, *iam.Options] {
	return &sources.GetListSource[*types.InstanceProfile, *iam.Client, *iam.Options]{
		ItemType:      "iam-instance-profile",
		Client:        client,
		CacheDuration: 3 * time.Hour, // IAM has very low rate limits, we need to cache for a long time
		AccountID:     accountID,
		GetFunc: func(ctx context.Context, client *iam.Client, scope, query string) (*types.InstanceProfile, error) {
			return instanceProfileGetFunc(ctx, client, scope, query)
		},
		ListFunc: func(ctx context.Context, client *iam.Client, scope string) ([]*types.InstanceProfile, error) {
			return instanceProfileListFunc(ctx, client, scope)
		},
		ListTagsFunc: func(ctx context.Context, ip *types.InstanceProfile, c *iam.Client) (map[string]string, error) {
			return instanceProfileListTagsFunc(ctx, ip, c), nil
		},
		ItemMapper: instanceProfileItemMapper,
	}
}

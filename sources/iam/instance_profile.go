package iam

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"

	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func instanceProfileGetFunc(ctx context.Context, client *iam.Client, scope, query string) (*types.InstanceProfile, error) {
	out, err := client.GetInstanceProfile(ctx, &iam.GetInstanceProfileInput{
		InstanceProfileName: &query,
	})

	if err != nil {
		return nil, err
	}

	return out.InstanceProfile, nil
}

func instanceProfileListFunc(ctx context.Context, client *iam.Client, scope string) ([]*types.InstanceProfile, error) {
	out, err := client.ListInstanceProfiles(ctx, &iam.ListInstanceProfilesInput{})

	if err != nil {
		return nil, err
	}

	zones := make([]*types.InstanceProfile, len(out.InstanceProfiles))

	for i := range out.InstanceProfiles {
		zones[i] = &out.InstanceProfiles[i]
	}

	return zones, nil
}

func instanceProfileItemMapper(scope string, awsItem *types.InstanceProfile) (*sdp.Item, error) {
	attributes, err := sources.ToAttributesCase(awsItem)

	if err != nil {
		return nil, err
	}

	item := sdp.Item{
		Type:            "iam-instance-profile",
		UniqueAttribute: "instanceProfileName",
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

//go:generate docgen ../../docs-data
// +overmind:type iam-instance-profile
// +overmind:descriptiveType IAM instance profile
// +overmind:get Get an IAM instance profile
// +overmind:list List IAM instance profiles
// +overmind:search Search IAM instance profiles by ARN
// +overmind:group AWS
// +overmind:terraform:queryMap aws_iam_instance_profile.arn
// +overmind:terraform:method SEARCH

func NewInstanceProfileSource(config aws.Config, accountID string, region string, limit *sources.LimitBucket) *sources.GetListSource[*types.InstanceProfile, *iam.Client, *iam.Options] {
	return &sources.GetListSource[*types.InstanceProfile, *iam.Client, *iam.Options]{
		ItemType:      "iam-instance-profile",
		Client:        iam.NewFromConfig(config),
		CacheDuration: 1 * time.Hour, // IAM has very low rate limits, we need to cache for a long time
		AccountID:     accountID,
		GetFunc: func(ctx context.Context, client *iam.Client, scope, query string) (*types.InstanceProfile, error) {
			<-limit.C
			return instanceProfileGetFunc(ctx, client, scope, query)
		},
		ListFunc: func(ctx context.Context, client *iam.Client, scope string) ([]*types.InstanceProfile, error) {
			<-limit.C
			return instanceProfileListFunc(ctx, client, scope)
		},
		ItemMapper: instanceProfileItemMapper,
	}
}
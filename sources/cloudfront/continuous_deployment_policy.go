package cloudfront

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func continuousDeploymentPolicyItemMapper(_, scope string, awsItem *types.ContinuousDeploymentPolicy) (*sdp.Item, error) {
	attributes, err := sources.ToAttributesWithExclude(awsItem)

	if err != nil {
		return nil, err
	}

	item := sdp.Item{
		Type:            "cloudfront-continuous-deployment-policy",
		UniqueAttribute: "Id",
		Attributes:      attributes,
		Scope:           scope,
	}

	if awsItem.ContinuousDeploymentPolicyConfig != nil && awsItem.ContinuousDeploymentPolicyConfig.StagingDistributionDnsNames != nil {
		for _, name := range awsItem.ContinuousDeploymentPolicyConfig.StagingDistributionDnsNames.Items {
			// +overmind:link dns
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "dns",
					Method: sdp.QueryMethod_SEARCH,
					Query:  name,
					Scope:  "global",
				},
				BlastPropagation: &sdp.BlastPropagation{
					// DNS is always linked
					In:  true,
					Out: true,
				},
			})
		}
	}

	return &item, nil
}

//go:generate docgen ../../docs-data
// +overmind:type cloudfront-continuous-deployment-policy
// +overmind:descriptiveType CloudFront Continuous Deployment Policy
// +overmind:get Get a CloudFront Continuous Deployment Policy by ID
// +overmind:list List CloudFront Continuous Deployment Policies
// +overmind:search Search CloudFront Continuous Deployment Policies by ARN
// +overmind:group AWS

// Terraform is not yet supported for this: https://github.com/hashicorp/terraform-provider-aws/issues/28920

func NewContinuousDeploymentPolicySource(client *cloudfront.Client, accountID string) *sources.GetListSource[*types.ContinuousDeploymentPolicy, *cloudfront.Client, *cloudfront.Options] {
	return &sources.GetListSource[*types.ContinuousDeploymentPolicy, *cloudfront.Client, *cloudfront.Options]{
		ItemType:               "cloudfront-continuous-deployment-policy",
		Client:                 client,
		AccountID:              accountID,
		Region:                 "",   // Cloudfront resources aren't tied to a region
		SupportGlobalResources: true, // Some policies are global
		GetFunc: func(ctx context.Context, client *cloudfront.Client, scope, query string) (*types.ContinuousDeploymentPolicy, error) {
			out, err := client.GetContinuousDeploymentPolicy(ctx, &cloudfront.GetContinuousDeploymentPolicyInput{
				Id: &query,
			})

			if err != nil {
				return nil, err
			}

			return out.ContinuousDeploymentPolicy, nil
		},
		ListFunc: func(ctx context.Context, client *cloudfront.Client, scope string) ([]*types.ContinuousDeploymentPolicy, error) {
			out, err := client.ListContinuousDeploymentPolicies(ctx, &cloudfront.ListContinuousDeploymentPoliciesInput{})

			if err != nil {
				return nil, err
			}

			policies := make([]*types.ContinuousDeploymentPolicy, 0, len(out.ContinuousDeploymentPolicyList.Items))

			for _, policy := range out.ContinuousDeploymentPolicyList.Items {
				policies = append(policies, policy.ContinuousDeploymentPolicy)
			}

			return policies, nil
		},
		ItemMapper: continuousDeploymentPolicyItemMapper,
	}
}

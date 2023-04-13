package eks

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func fargateProfileGetFunc(ctx context.Context, client EKSClient, scope string, input *eks.DescribeFargateProfileInput) (*sdp.Item, error) {
	out, err := client.DescribeFargateProfile(ctx, input)

	if err != nil {
		return nil, err
	}

	if out.FargateProfile == nil {
		return nil, &sdp.QueryError{
			ErrorType:   sdp.QueryError_NOTFOUND,
			ErrorString: "fargate profile was nil",
		}
	}

	attributes, err := sources.ToAttributesCase(out.FargateProfile)

	if err != nil {
		return nil, err
	}

	// The uniqueAttributeValue for this is a custom field:
	// {clusterName}/{FargateProfileName}
	attributes.Set("uniqueName", (*out.FargateProfile.ClusterName + "/" + *out.FargateProfile.FargateProfileName))

	item := sdp.Item{
		Type:            "eks-fargate-profile",
		UniqueAttribute: "uniqueName",
		Attributes:      attributes,
		Scope:           scope,
	}

	if out.FargateProfile.PodExecutionRoleArn != nil {
		if a, err := sources.ParseARN(*out.FargateProfile.PodExecutionRoleArn); err == nil {
			// +overmind:link iam-role
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
				Type:   "iam-role",
				Method: sdp.QueryMethod_SEARCH,
				Query:  *out.FargateProfile.PodExecutionRoleArn,
				Scope:  sources.FormatScope(a.AccountID, a.Region),
			})
		}
	}

	for _, subnet := range out.FargateProfile.Subnets {
		// +overmind:link ec2-subnet
		item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
			Type:   "ec2-subnet",
			Method: sdp.QueryMethod_GET,
			Query:  subnet,
			Scope:  scope,
		})
	}

	return &item, nil
}

//go:generate docgen ../../docs-data
// +overmind:type eks-fargate-profile
// +overmind:descriptiveType Fargate Profile
// +overmind:get Get a fargate profile by unique name ({clusterName}/{FargateProfileName})
// +overmind:list List all fargate profiles
// +overmind:search Search for fargate profiles by cluster name
// +overmind:group AWS

func NewFargateProfileSource(config aws.Config, accountID string, region string) *sources.AlwaysGetSource[*eks.ListFargateProfilesInput, *eks.ListFargateProfilesOutput, *eks.DescribeFargateProfileInput, *eks.DescribeFargateProfileOutput, EKSClient, *eks.Options] {
	return &sources.AlwaysGetSource[*eks.ListFargateProfilesInput, *eks.ListFargateProfilesOutput, *eks.DescribeFargateProfileInput, *eks.DescribeFargateProfileOutput, EKSClient, *eks.Options]{
		ItemType:    "eks-fargate-profile",
		Client:      eks.NewFromConfig(config),
		AccountID:   accountID,
		Region:      region,
		DisableList: true,
		SearchInputMapper: func(scope, query string) (*eks.ListFargateProfilesInput, error) {
			return &eks.ListFargateProfilesInput{
				ClusterName: &query,
			}, nil
		},
		GetInputMapper: func(scope, query string) *eks.DescribeFargateProfileInput {
			// The uniqueAttributeValue for this is a custom field:
			// {clusterName}/{FargateProfileName}
			fields := strings.Split(query, "/")

			var clusterName string
			var FargateProfileName string

			if len(fields) == 2 {
				clusterName = fields[0]
				FargateProfileName = fields[1]
			}

			return &eks.DescribeFargateProfileInput{
				FargateProfileName: &FargateProfileName,
				ClusterName:        &clusterName,
			}
		},
		ListFuncPaginatorBuilder: func(client EKSClient, input *eks.ListFargateProfilesInput) sources.Paginator[*eks.ListFargateProfilesOutput, *eks.Options] {
			return eks.NewListFargateProfilesPaginator(client, input)
		},
		ListFuncOutputMapper: func(output *eks.ListFargateProfilesOutput, input *eks.ListFargateProfilesInput) ([]*eks.DescribeFargateProfileInput, error) {
			inputs := make([]*eks.DescribeFargateProfileInput, len(output.FargateProfileNames))

			for i := range output.FargateProfileNames {
				inputs[i] = &eks.DescribeFargateProfileInput{
					ClusterName:        input.ClusterName,
					FargateProfileName: &output.FargateProfileNames[i],
				}
			}

			return inputs, nil
		},
		GetFunc: fargateProfileGetFunc,
	}
}

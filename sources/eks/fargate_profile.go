package eks

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func FargateProfileGetFunc(ctx context.Context, client EKSClient, scope string, input *eks.DescribeFargateProfileInput) (*sdp.Item, error) {
	out, err := client.DescribeFargateProfile(ctx, input)

	if err != nil {
		return nil, err
	}

	if out.FargateProfile == nil {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOTFOUND,
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
			item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
				Type:   "iam-role",
				Method: sdp.RequestMethod_SEARCH,
				Query:  *out.FargateProfile.PodExecutionRoleArn,
				Scope:  sources.FormatScope(a.AccountID, a.Region),
			})
		}
	}

	return &item, nil
}

func NewFargateProfileSource(config aws.Config, accountID string, region string) *sources.ListGetSource[*eks.ListFargateProfilesInput, *eks.ListFargateProfilesOutput, *eks.DescribeFargateProfileInput, *eks.DescribeFargateProfileOutput, EKSClient, *eks.Options] {
	return &sources.ListGetSource[*eks.ListFargateProfilesInput, *eks.ListFargateProfilesOutput, *eks.DescribeFargateProfileInput, *eks.DescribeFargateProfileOutput, EKSClient, *eks.Options]{
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

			for i, name := range output.FargateProfileNames {
				inputs[i] = &eks.DescribeFargateProfileInput{
					ClusterName:        input.ClusterName,
					FargateProfileName: &name,
				}
			}

			return inputs, nil
		},
		GetFunc: FargateProfileGetFunc,
	}
}

package sns

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

type platformApplicationClient interface {
	ListPlatformApplications(ctx context.Context, params *sns.ListPlatformApplicationsInput, optFns ...func(*sns.Options)) (*sns.ListPlatformApplicationsOutput, error)
	GetPlatformApplicationAttributes(ctx context.Context, params *sns.GetPlatformApplicationAttributesInput, optFns ...func(*sns.Options)) (*sns.GetPlatformApplicationAttributesOutput, error)
	ListTagsForResource(context.Context, *sns.ListTagsForResourceInput, ...func(*sns.Options)) (*sns.ListTagsForResourceOutput, error)
}

func getPlatformApplicationFunc(ctx context.Context, client platformApplicationClient, scope string, input *sns.GetPlatformApplicationAttributesInput) (*sdp.Item, error) {
	output, err := client.GetPlatformApplicationAttributes(ctx, input)
	if err != nil {
		return nil, err
	}

	if output.Attributes == nil {
		return nil, &sdp.QueryError{
			ErrorType:   sdp.QueryError_NOTFOUND,
			ErrorString: "get platform application attributes response was nil",
		}
	}

	attributes, err := sources.ToAttributesCase(output.Attributes)
	if err != nil {
		return nil, err
	}

	err = attributes.Set("platformApplicationArn", *input.PlatformApplicationArn)
	if err != nil {
		return nil, err
	}

	item := &sdp.Item{
		Type:            "sns-platform-application",
		UniqueAttribute: "platformApplicationArn",
		Attributes:      attributes,
		Scope:           scope,
	}

	if resourceTags, err := tagsByResourceARN(ctx, client, *input.PlatformApplicationArn); err == nil {
		item.Tags = tagsToMap(resourceTags)
	}

	// +overmind:link sns-endpoint
	item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
		Query: &sdp.Query{
			Type:   "sns-endpoint",
			Method: sdp.QueryMethod_SEARCH,
			Query:  *input.PlatformApplicationArn,
			Scope:  scope,
		},
		BlastPropagation: &sdp.BlastPropagation{
			// An unhealthy endpoint won't affect the platform application
			In: false,
			// If platform application is unhealthy, then endpoints won't get notifications
			Out: true,
		},
	})

	return item, nil
}

//go:generate docgen ../../docs-data
// +overmind:type sns-platform-application
// +overmind:descriptiveType SNS Platform Application
// +overmind:get Get an SNS platform application by its ARN
// +overmind:list List all SNS platform applications
// +overmind:search Search SNS platform applications by ARN
// +overmind:group AWS
// +overmind:terraform:queryMap aws_sns_platform_application.id

func NewPlatformApplicationSource(config aws.Config, accountID string, region string) *sources.AlwaysGetSource[*sns.ListPlatformApplicationsInput, *sns.ListPlatformApplicationsOutput, *sns.GetPlatformApplicationAttributesInput, *sns.GetPlatformApplicationAttributesOutput, platformApplicationClient, *sns.Options] {
	return &sources.AlwaysGetSource[*sns.ListPlatformApplicationsInput, *sns.ListPlatformApplicationsOutput, *sns.GetPlatformApplicationAttributesInput, *sns.GetPlatformApplicationAttributesOutput, platformApplicationClient, *sns.Options]{
		ItemType:  "sns-platform-application",
		Client:    sns.NewFromConfig(config),
		AccountID: accountID,
		Region:    region,
		ListInput: &sns.ListPlatformApplicationsInput{},
		GetInputMapper: func(scope, query string) *sns.GetPlatformApplicationAttributesInput {
			return &sns.GetPlatformApplicationAttributesInput{
				PlatformApplicationArn: &query,
			}
		},
		ListFuncPaginatorBuilder: func(client platformApplicationClient, input *sns.ListPlatformApplicationsInput) sources.Paginator[*sns.ListPlatformApplicationsOutput, *sns.Options] {
			return sns.NewListPlatformApplicationsPaginator(client, input)
		},
		ListFuncOutputMapper: func(output *sns.ListPlatformApplicationsOutput, input *sns.ListPlatformApplicationsInput) ([]*sns.GetPlatformApplicationAttributesInput, error) {
			var inputs []*sns.GetPlatformApplicationAttributesInput
			for _, platformApplication := range output.PlatformApplications {
				inputs = append(inputs, &sns.GetPlatformApplicationAttributesInput{
					PlatformApplicationArn: platformApplication.PlatformApplicationArn,
				})
			}
			return inputs, nil
		},
		GetFunc: getPlatformApplicationFunc,
	}
}

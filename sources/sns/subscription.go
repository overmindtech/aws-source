package sns

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

type subsCli interface {
	GetSubscriptionAttributes(ctx context.Context, params *sns.GetSubscriptionAttributesInput, optFns ...func(*sns.Options)) (*sns.GetSubscriptionAttributesOutput, error)
	ListSubscriptions(context.Context, *sns.ListSubscriptionsInput, ...func(*sns.Options)) (*sns.ListSubscriptionsOutput, error)
	ListTagsForResource(context.Context, *sns.ListTagsForResourceInput, ...func(*sns.Options)) (*sns.ListTagsForResourceOutput, error)
}

func getSubsFunc(ctx context.Context, client subsCli, scope string, input *sns.GetSubscriptionAttributesInput) (*sdp.Item, error) {
	output, err := client.GetSubscriptionAttributes(ctx, input)
	if err != nil {
		return nil, err
	}

	if output.Attributes == nil {
		return nil, &sdp.QueryError{
			ErrorType:   sdp.QueryError_NOTFOUND,
			ErrorString: "get subscription attributes response was nil",
		}
	}

	attributes, err := sources.ToAttributesCase(output.Attributes)
	if err != nil {
		return nil, err
	}

	item := &sdp.Item{
		Type:            "sns-subscription",
		UniqueAttribute: "subscriptionArn",
		Attributes:      attributes,
		Scope:           scope,
	}

	if resourceTags, err := tagsByResourceARN(ctx, client, *input.SubscriptionArn); err == nil {
		item.Tags = tagsToMap(resourceTags)
	}

	if topicArn, err := attributes.Get("topicArn"); err == nil {
		// +overmind:link sns-topic
		item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
			Query: &sdp.Query{
				Type:   "sns-topic",
				Method: sdp.QueryMethod_GET,
				Query:  topicArn.(string),
				Scope:  scope,
			},
			BlastPropagation: &sdp.BlastPropagation{
				// If topic is not healthy, subscription will not work
				In: true,
				// Subscription won't affect the topic
				Out: false,
			},
		})
	}

	if subsRoleArn, err := attributes.Get("subscriptionRoleArn"); err == nil {
		if arn, err := sources.ParseARN(fmt.Sprint(subsRoleArn)); err == nil {
			// +overmind:link iam-role
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "iam-role",
					Method: sdp.QueryMethod_GET,
					Query:  arn.ResourceID(),
					Scope:  sources.FormatScope(arn.AccountID, arn.Region),
				},
				BlastPropagation: &sdp.BlastPropagation{
					// If role is not healthy, subscription will not work
					In: true,
					// Subscription won't affect the role
					Out: false,
				},
			})
		}
	}

	return item, nil
}

//go:generate docgen ../../docs-data
// +overmind:type sns-subscription
// +overmind:descriptiveType SNS Subscription
// +overmind:get Get an SNS subscription by its ARN
// +overmind:list List all SNS subscriptions
// +overmind:search Search SNS subscription by ARN
// +overmind:group AWS
// +overmind:terraform:queryMap aws_sns_topic_subscription.id

func NewSubscriptionSource(config aws.Config, accountID string, region string) *sources.AlwaysGetSource[*sns.ListSubscriptionsInput, *sns.ListSubscriptionsOutput, *sns.GetSubscriptionAttributesInput, *sns.GetSubscriptionAttributesOutput, subsCli, *sns.Options] {
	return &sources.AlwaysGetSource[*sns.ListSubscriptionsInput, *sns.ListSubscriptionsOutput, *sns.GetSubscriptionAttributesInput, *sns.GetSubscriptionAttributesOutput, subsCli, *sns.Options]{
		ItemType:  "sns-subscription",
		Client:    sns.NewFromConfig(config),
		AccountID: accountID,
		Region:    region,
		ListInput: &sns.ListSubscriptionsInput{},
		GetInputMapper: func(scope, query string) *sns.GetSubscriptionAttributesInput {
			return &sns.GetSubscriptionAttributesInput{
				SubscriptionArn: &query,
			}
		},
		ListFuncPaginatorBuilder: func(client subsCli, input *sns.ListSubscriptionsInput) sources.Paginator[*sns.ListSubscriptionsOutput, *sns.Options] {
			return sns.NewListSubscriptionsPaginator(client, input)
		},
		ListFuncOutputMapper: func(output *sns.ListSubscriptionsOutput, _ *sns.ListSubscriptionsInput) ([]*sns.GetSubscriptionAttributesInput, error) {
			var inputs []*sns.GetSubscriptionAttributesInput
			for _, subs := range output.Subscriptions {
				inputs = append(inputs, &sns.GetSubscriptionAttributesInput{
					SubscriptionArn: subs.SubscriptionArn,
				})
			}
			return inputs, nil
		},
		GetFunc: getSubsFunc,
	}
}

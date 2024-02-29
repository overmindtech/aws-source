package sns

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/overmindtech/aws-source/sources"
)

type testTopicClient struct{}

func (t testTopicClient) GetTopicAttributes(ctx context.Context, params *sns.GetTopicAttributesInput, optFns ...func(*sns.Options)) (*sns.GetTopicAttributesOutput, error) {
	return &sns.GetTopicAttributesOutput{Attributes: map[string]string{
		"SubscriptionsConfirmed":  "1",
		"DisplayName":             "my-topic",
		"SubscriptionsDeleted":    "0",
		"EffectiveDeliveryPolicy": "{\"http\":{\"defaultHealthyRetryPolicy\":{\"minDelayTarget\":20,\"maxDelayTarget\":20,\"numRetries\":3,\"numMaxDelayRetries\":0,\"numNoDelayRetries\":0,\"numMinDelayRetries\":0,\"backoffFunction\":\"linear\"},\"disableSubscriptionOverrides\":false}}",
		"Owner":                   "123456789012",
		"Policy":                  "{\"Version\":\"2008-10-17\",\"Id\":\"__default_policy_ID\",\"Statement\":[{\"Sid\":\"__default_statement_ID\",\"Effect\":\"Allow\",\"Principal\":{\"AWS\":\"*\"},\"Action\":[\"SNS:Subscribe\",\"SNS:ListSubscriptionsByTopic\",\"SNS:DeleteTopic\",\"SNS:GetTopicAttributes\",\"SNS:Publish\",\"SNS:RemovePermission\",\"SNS:AddPermission\",\"SNS:SetTopicAttributes\"],\"Resource\":\"arn:aws:sns:us-west-2:123456789012:my-topic\",\"Condition\":{\"StringEquals\":{\"AWS:SourceOwner\":\"0123456789012\"}}}]}",
		"TopicArn":                "arn:aws:sns:us-west-2:123456789012:my-topic",
		"SubscriptionsPending":    "0",
		"KmsMasterKeyId":          "alias/aws/sns",
	}}, nil
}

func (t testTopicClient) ListTopics(context.Context, *sns.ListTopicsInput, ...func(*sns.Options)) (*sns.ListTopicsOutput, error) {
	return &sns.ListTopicsOutput{
		Topics: []types.Topic{
			{
				TopicArn: sources.PtrString("arn:aws:sns:us-west-2:123456789012:my-topic"),
			},
		},
	}, nil
}

func (t testTopicClient) ListTagsForResource(context.Context, *sns.ListTagsForResourceInput, ...func(*sns.Options)) (*sns.ListTagsForResourceOutput, error) {
	return &sns.ListTagsForResourceOutput{
		Tags: []types.Tag{
			{Key: sources.PtrString("tag1"), Value: sources.PtrString("value1")},
			{Key: sources.PtrString("tag2"), Value: sources.PtrString("value2")},
		},
	}, nil
}

func TestGetTopicFunc(t *testing.T) {
	ctx := context.Background()
	cli := testTopicClient{}

	item, err := getTopicFunc(ctx, cli, "scope", &sns.GetTopicAttributesInput{
		TopicArn: sources.PtrString("arn:aws:sns:us-west-2:123456789012:my-topic"),
	})
	if err != nil {
		t.Fatal(err)
	}

	if err = item.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestNewTopicSource(t *testing.T) {
	config, account, region := sources.GetAutoConfig(t)

	source := NewTopicSource(config, account, region)

	test := sources.E2ETest{
		Source:  source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}

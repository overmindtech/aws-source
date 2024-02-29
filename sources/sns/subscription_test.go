package sns

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/overmindtech/aws-source/sources"
)

type testClient struct{}

func (t testClient) GetSubscriptionAttributes(ctx context.Context, params *sns.GetSubscriptionAttributesInput, optFns ...func(*sns.Options)) (*sns.GetSubscriptionAttributesOutput, error) {
	return &sns.GetSubscriptionAttributesOutput{Attributes: map[string]string{
		"Endpoint":                     "my-email@example.com",
		"Protocol":                     "email",
		"RawMessageDelivery":           "false",
		"ConfirmationWasAuthenticated": "false",
		"Owner":                        "123456789012",
		"SubscriptionArn":              "arn:aws:sns:us-west-2:123456789012:my-topic:8a21d249-4329-4871-acc6-7be709c6ea7f",
		"TopicArn":                     "arn:aws:sns:us-west-2:123456789012:my-topic",
		"SubscriptionRoleArn":          "arn:aws:iam::123456789012:role/my-role",
	}}, nil
}

func (t testClient) ListSubscriptions(context.Context, *sns.ListSubscriptionsInput, ...func(*sns.Options)) (*sns.ListSubscriptionsOutput, error) {
	return &sns.ListSubscriptionsOutput{
		Subscriptions: []types.Subscription{
			{
				Owner:           sources.PtrString("123456789012"),
				Endpoint:        sources.PtrString("my-email@example.com"),
				Protocol:        sources.PtrString("email"),
				TopicArn:        sources.PtrString("arn:aws:sns:us-west-2:123456789012:my-topic"),
				SubscriptionArn: sources.PtrString("arn:aws:sns:us-west-2:123456789012:my-topic:8a21d249-4329-4871-acc6-7be709c6ea7f"),
			},
		},
	}, nil
}

func (t testClient) ListTagsForResource(context.Context, *sns.ListTagsForResourceInput, ...func(*sns.Options)) (*sns.ListTagsForResourceOutput, error) {
	return &sns.ListTagsForResourceOutput{
		Tags: []types.Tag{
			{Key: sources.PtrString("tag1"), Value: sources.PtrString("value1")},
			{Key: sources.PtrString("tag2"), Value: sources.PtrString("value2")},
		},
	}, nil
}

func TestGetFunc(t *testing.T) {
	ctx := context.Background()
	cli := testClient{}

	item, err := getSubsFunc(ctx, cli, "scope", &sns.GetSubscriptionAttributesInput{
		SubscriptionArn: sources.PtrString("arn:aws:sns:us-west-2:123456789012:my-topic:8a21d249-4329-4871-acc6-7be709c6ea7f"),
	})
	if err != nil {
		t.Fatal(err)
	}

	if err = item.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestNewSubscriptionSource(t *testing.T) {
	config, account, region := sources.GetAutoConfig(t)

	source := NewSubscriptionSource(config, account, region)

	test := sources.E2ETest{
		Source:  source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}

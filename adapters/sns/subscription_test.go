package sns

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/overmindtech/aws-source/adapters"
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
				Owner:           adapters.PtrString("123456789012"),
				Endpoint:        adapters.PtrString("my-email@example.com"),
				Protocol:        adapters.PtrString("email"),
				TopicArn:        adapters.PtrString("arn:aws:sns:us-west-2:123456789012:my-topic"),
				SubscriptionArn: adapters.PtrString("arn:aws:sns:us-west-2:123456789012:my-topic:8a21d249-4329-4871-acc6-7be709c6ea7f"),
			},
		},
	}, nil
}

func (t testClient) ListTagsForResource(context.Context, *sns.ListTagsForResourceInput, ...func(*sns.Options)) (*sns.ListTagsForResourceOutput, error) {
	return &sns.ListTagsForResourceOutput{
		Tags: []types.Tag{
			{Key: adapters.PtrString("tag1"), Value: adapters.PtrString("value1")},
			{Key: adapters.PtrString("tag2"), Value: adapters.PtrString("value2")},
		},
	}, nil
}

func TestGetFunc(t *testing.T) {
	ctx := context.Background()
	cli := testClient{}

	item, err := getSubsFunc(ctx, cli, "scope", &sns.GetSubscriptionAttributesInput{
		SubscriptionArn: adapters.PtrString("arn:aws:sns:us-west-2:123456789012:my-topic:8a21d249-4329-4871-acc6-7be709c6ea7f"),
	})
	if err != nil {
		t.Fatal(err)
	}

	if err = item.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestNewSubscriptionSource(t *testing.T) {
	client, account, region := GetAutoConfig(t)

	source := NewSubscriptionSource(client, account, region)

	test := adapters.E2ETest{
		Adapter: source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}

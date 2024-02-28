package sqs

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/overmindtech/aws-source/sources"
)

type testClient struct{}

func (t testClient) GetQueueAttributes(ctx context.Context, params *sqs.GetQueueAttributesInput, optFns ...func(*sqs.Options)) (*sqs.GetQueueAttributesOutput, error) {
	return &sqs.GetQueueAttributesOutput{
		Attributes: map[string]string{
			"ApproximateNumberOfMessages":           "0",
			"ApproximateNumberOfMessagesDelayed":    "0",
			"ApproximateNumberOfMessagesNotVisible": "0",
			"CreatedTimestamp":                      "1631616000",
			"DelaySeconds":                          "0",
			"LastModifiedTimestamp":                 "1631616000",
			"MaximumMessageSize":                    "262144",
			"MessageRetentionPeriod":                "345600",
			"QueueArn":                              "arn:aws:sqs:us-west-2:123456789012:MyQueue",
			"ReceiveMessageWaitTimeSeconds":         "0",
			"VisibilityTimeout":                     "30",
			"RedrivePolicy":                         "{\"deadLetterTargetArn\":\"arn:aws:sqs:us-east-1:80398EXAMPLE:MyDeadLetterQueue\",\"maxReceiveCount\":1000}",
		},
	}, nil
}

func (t testClient) ListQueueTags(ctx context.Context, params *sqs.ListQueueTagsInput, optFns ...func(*sqs.Options)) (*sqs.ListQueueTagsOutput, error) {
	return &sqs.ListQueueTagsOutput{
		Tags: map[string]string{
			"tag1": "value1",
			"tag2": "value2",
		},
	}, nil
}

func (t testClient) ListQueues(ctx context.Context, input *sqs.ListQueuesInput, f ...func(*sqs.Options)) (*sqs.ListQueuesOutput, error) {
	return &sqs.ListQueuesOutput{
		QueueUrls: []string{
			"https://sqs.us-west-2.amazonaws.com/123456789012/MyQueue",
			"https://sqs.us-west-2.amazonaws.com/123456789012/MyQueue2",
		},
	}, nil
}

func TestGetFunc(t *testing.T) {
	ctx := context.Background()
	cli := testClient{}

	item, err := getFunc(ctx, cli, "scope", &sqs.GetQueueAttributesInput{
		QueueUrl: sources.PtrString("https://sqs.us-west-2.amazonaws.com/123456789012/MyQueue"),
	})
	if err != nil {
		t.Fatal(err)
	}

	if err = item.Validate(); err != nil {
		t.Error(err)
	}
}

func TestNewQueueSource(t *testing.T) {
	config, account, region := sources.GetAutoConfig(t)

	source := NewQueueSource(config, account, region)

	test := sources.E2ETest{
		Source:  source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}

package sqs

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

type client interface {
	GetQueueAttributes(ctx context.Context, params *sqs.GetQueueAttributesInput, optFns ...func(*sqs.Options)) (*sqs.GetQueueAttributesOutput, error)
	ListQueueTags(ctx context.Context, params *sqs.ListQueueTagsInput, optFns ...func(*sqs.Options)) (*sqs.ListQueueTagsOutput, error)
	ListQueues(context.Context, *sqs.ListQueuesInput, ...func(*sqs.Options)) (*sqs.ListQueuesOutput, error)
}

func getFunc(ctx context.Context, client client, scope string, input *sqs.GetQueueAttributesInput) (*sdp.Item, error) {
	output, err := client.GetQueueAttributes(ctx, input)
	if err != nil {
		return nil, err
	}

	if output.Attributes == nil {
		return nil, &sdp.QueryError{
			ErrorType:   sdp.QueryError_NOTFOUND,
			ErrorString: "get queue attributes response was nil",
		}
	}

	attributes, err := sources.ToAttributesCase(output.Attributes)
	if err != nil {
		return nil, err
	}

	err = attributes.Set("queueURL", input.QueueUrl)
	if err != nil {
		return nil, err
	}

	resourceTags, err := tags(ctx, client, *input.QueueUrl)
	if err != nil {
		return nil, &sdp.QueryError{
			ErrorType:   sdp.QueryError_NOTFOUND,
			ErrorString: err.Error(),
		}
	}

	return &sdp.Item{
		Type:            "sqs-queue",
		UniqueAttribute: "queueURL",
		Attributes:      attributes,
		Scope:           scope,
		Tags:            resourceTags,
	}, nil
}

//go:generate docgen ../../docs-data
// +overmind:type sqs-queue
// +overmind:descriptiveType SQS Queue
// +overmind:get Get an SQS queue attributes by its URL
// +overmind:list List all SQS queue URLs
// +overmind:search Search SQS queue by ARN
// +overmind:group AWS
// +overmind:terraform:queryMap aws_sqs_queue.id

func NewQueueSource(config aws.Config, accountID string, region string) *sources.AlwaysGetSource[*sqs.ListQueuesInput, *sqs.ListQueuesOutput, *sqs.GetQueueAttributesInput, *sqs.GetQueueAttributesOutput, client, *sqs.Options] {
	return &sources.AlwaysGetSource[*sqs.ListQueuesInput, *sqs.ListQueuesOutput, *sqs.GetQueueAttributesInput, *sqs.GetQueueAttributesOutput, client, *sqs.Options]{
		ItemType:  "sqs-queue",
		Client:    sqs.NewFromConfig(config),
		AccountID: accountID,
		Region:    region,
		ListInput: &sqs.ListQueuesInput{},
		GetInputMapper: func(scope, query string) *sqs.GetQueueAttributesInput {
			return &sqs.GetQueueAttributesInput{
				QueueUrl: &query,
				// Providing All will return all attributes.
				AttributeNames: []types.QueueAttributeName{"All"},
			}
		},
		ListFuncPaginatorBuilder: func(client client, input *sqs.ListQueuesInput) sources.Paginator[*sqs.ListQueuesOutput, *sqs.Options] {
			return sqs.NewListQueuesPaginator(client, input)
		},
		ListFuncOutputMapper: func(output *sqs.ListQueuesOutput, _ *sqs.ListQueuesInput) ([]*sqs.GetQueueAttributesInput, error) {
			var inputs []*sqs.GetQueueAttributesInput
			for _, url := range output.QueueUrls {
				inputs = append(inputs, &sqs.GetQueueAttributesInput{
					QueueUrl: &url,
				})
			}
			return inputs, nil
		},
		GetFunc: getFunc,
	}
}

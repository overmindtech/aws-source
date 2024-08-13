package kms

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/kms/types"

	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

type kmsClient interface {
	DescribeKey(ctx context.Context, params *kms.DescribeKeyInput, optFns ...func(*kms.Options)) (*kms.DescribeKeyOutput, error)
	ListKeys(context.Context, *kms.ListKeysInput, ...func(*kms.Options)) (*kms.ListKeysOutput, error)
	ListResourceTags(context.Context, *kms.ListResourceTagsInput, ...func(*kms.Options)) (*kms.ListResourceTagsOutput, error)
}

func getFunc(ctx context.Context, client kmsClient, scope string, input *kms.DescribeKeyInput) (*sdp.Item, error) {
	output, err := client.DescribeKey(ctx, input)
	if err != nil {
		return nil, err
	}

	if output.KeyMetadata == nil {
		return nil, &sdp.QueryError{
			ErrorType:   sdp.QueryError_NOTFOUND,
			ErrorString: "describe key response was nil",
		}
	}

	attributes, err := sources.ToAttributesCase(output.KeyMetadata)
	if err != nil {
		return nil, err
	}

	// Some keys can be accessed, but not their tags, even if you have full
	// admin access. No clue how this is possible but seems to be an
	// inconsistency in the AWS API. In this case, we will ignore the error and
	// embed it in a tag so that you can see that they are missing
	var resourceTags map[string]string
	resourceTags, err = tags(ctx, client, *input.KeyId)
	if err != nil {
		resourceTags = sources.HandleTagsError(ctx, err)
	}

	item := &sdp.Item{
		Type:            "kms-key",
		UniqueAttribute: "keyId",
		Attributes:      attributes,
		Scope:           scope,
		Tags:            resourceTags,
	}

	if output.KeyMetadata.CustomKeyStoreId != nil {
		// +overmind:link kms-custom-key-store
		item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
			Query: &sdp.Query{
				Type:   "kms-custom-key-store",
				Method: sdp.QueryMethod_GET,
				Query:  *output.KeyMetadata.CustomKeyStoreId,
				Scope:  scope,
			},
			BlastPropagation: &sdp.BlastPropagation{
				// A keystore cannot be deleted if it contains a key.
				In: true,
				// Any change on the key won't affect the keystore.
				Out: false,
			},
		})
	}

	switch output.KeyMetadata.KeyState {
	case types.KeyStateEnabled:
		item.Health = sdp.Health_HEALTH_OK.Enum()
	case types.KeyStateUnavailable, types.KeyStateDisabled:
		item.Health = sdp.Health_HEALTH_UNKNOWN.Enum()
	case types.KeyStateCreating,
		types.KeyStatePendingDeletion,
		types.KeyStatePendingReplicaDeletion,
		types.KeyStatePendingImport,
		types.KeyStateUpdating:
		item.Health = sdp.Health_HEALTH_PENDING.Enum()
	default:
		return nil, &sdp.QueryError{
			ErrorType:   sdp.QueryError_OTHER,
			ErrorString: "unknown Key State",
		}
	}

	return item, nil
}

//go:generate docgen ../../docs-data
// +overmind:type kms-key
// +overmind:descriptiveType KMS Key
// +overmind:get Get a KMS Key by its ID
// +overmind:list List all KMS Keys
// +overmind:search Search for KMS Keys by ARN
// +overmind:group AWS
// +overmind:terraform:queryMap aws_kms_key.key_id

func NewKeySource(client kmsClient, accountID, region string) *sources.AlwaysGetSource[*kms.ListKeysInput, *kms.ListKeysOutput, *kms.DescribeKeyInput, *kms.DescribeKeyOutput, kmsClient, *kms.Options] {
	return &sources.AlwaysGetSource[*kms.ListKeysInput, *kms.ListKeysOutput, *kms.DescribeKeyInput, *kms.DescribeKeyOutput, kmsClient, *kms.Options]{
		ItemType:  "kms-key",
		Client:    client,
		AccountID: accountID,
		Region:    region,
		ListInput: &kms.ListKeysInput{},
		GetInputMapper: func(scope, query string) *kms.DescribeKeyInput {
			return &kms.DescribeKeyInput{
				KeyId: &query,
			}
		},
		ListFuncPaginatorBuilder: func(client kmsClient, input *kms.ListKeysInput) sources.Paginator[*kms.ListKeysOutput, *kms.Options] {
			return kms.NewListKeysPaginator(client, input)
		},
		ListFuncOutputMapper: func(output *kms.ListKeysOutput, _ *kms.ListKeysInput) ([]*kms.DescribeKeyInput, error) {
			var inputs []*kms.DescribeKeyInput
			for _, key := range output.Keys {
				inputs = append(inputs, &kms.DescribeKeyInput{
					KeyId: key.KeyId,
				})
			}
			return inputs, nil
		},
		GetFunc: getFunc,
	}
}

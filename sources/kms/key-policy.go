package kms

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/micahhausler/aws-iam-policy/policy"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/aws-source/sources/iam"
	"github.com/overmindtech/sdp-go"

	log "github.com/sirupsen/logrus"
)

type keyPolicyClient interface {
	GetKeyPolicy(ctx context.Context, params *kms.GetKeyPolicyInput, optFns ...func(*kms.Options)) (*kms.GetKeyPolicyOutput, error)
	ListKeyPolicies(ctx context.Context, params *kms.ListKeyPoliciesInput, optFns ...func(*kms.Options)) (*kms.ListKeyPoliciesOutput, error)
}

func getKeyPolicyFunc(ctx context.Context, client keyPolicyClient, scope string, input *kms.GetKeyPolicyInput) (*sdp.Item, error) {
	output, err := client.GetKeyPolicy(ctx, input)
	if err != nil {
		return nil, err
	}

	if output.Policy == nil {
		return nil, &sdp.QueryError{
			ErrorType:   sdp.QueryError_NOTFOUND,
			ErrorString: "get key policy response was nil",
		}
	}

	type keyParsedPolicy struct {
		*kms.GetKeyPolicyOutput
		PolicyDocument *policy.Policy
	}

	parsedPolicy := keyParsedPolicy{
		GetKeyPolicyOutput: output,
	}

	parsedPolicy.PolicyDocument, err = iam.ParsePolicyDocument(*output.Policy)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"input": input,
			"scope": scope,
		}).Error("Error parsing policy document")

		return nil, nil //nolint:nilerr
	}

	attributes, err := sources.ToAttributesWithExclude(parsedPolicy)
	if err != nil {
		return nil, err
	}

	err = attributes.Set("KeyId", *input.KeyId)
	if err != nil {
		return nil, err
	}

	item := &sdp.Item{
		Type:            "kms-key-policy",
		UniqueAttribute: "KeyId",
		Attributes:      attributes,
		Scope:           scope,
	}

	// +overmind:link kms-key
	item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
		Query: &sdp.Query{
			Type:   "kms-key",
			Method: sdp.QueryMethod_GET,
			Query:  *input.KeyId,
			Scope:  scope,
		},
		BlastPropagation: &sdp.BlastPropagation{
			// These are tightly coupled
			In:  true,
			Out: true,
		},
	})

	return item, nil
}

//go:generate docgen ../../docs-data
// +overmind:type kms-key-policy
// +overmind:descriptiveType KMS Key Policy
// +overmind:get Get a KMS key policy by its Key ID
// +overmind:search Search KMS key policies by Key ID
// +overmind:group AWS
// +overmind:terraform:queryMap aws_kms_key_policy.key_id

func NewKeyPolicySource(client keyPolicyClient, accountID string, region string) *sources.AlwaysGetSource[*kms.ListKeyPoliciesInput, *kms.ListKeyPoliciesOutput, *kms.GetKeyPolicyInput, *kms.GetKeyPolicyOutput, keyPolicyClient, *kms.Options] {
	return &sources.AlwaysGetSource[*kms.ListKeyPoliciesInput, *kms.ListKeyPoliciesOutput, *kms.GetKeyPolicyInput, *kms.GetKeyPolicyOutput, keyPolicyClient, *kms.Options]{
		ItemType:    "kms-key-policy",
		Client:      client,
		AccountID:   accountID,
		Region:      region,
		DisableList: true, // This source only supports listing by Key ID
		SearchInputMapper: func(scope, query string) (*kms.ListKeyPoliciesInput, error) {
			return &kms.ListKeyPoliciesInput{
				KeyId: &query,
			}, nil
		},
		GetInputMapper: func(scope, query string) *kms.GetKeyPolicyInput {
			return &kms.GetKeyPolicyInput{
				KeyId: &query,
			}
		},
		ListFuncPaginatorBuilder: func(client keyPolicyClient, input *kms.ListKeyPoliciesInput) sources.Paginator[*kms.ListKeyPoliciesOutput, *kms.Options] {
			return kms.NewListKeyPoliciesPaginator(client, input)
		},
		ListFuncOutputMapper: func(output *kms.ListKeyPoliciesOutput, input *kms.ListKeyPoliciesInput) ([]*kms.GetKeyPolicyInput, error) {
			var inputs []*kms.GetKeyPolicyInput
			for _, policyName := range output.PolicyNames {
				inputs = append(inputs, &kms.GetKeyPolicyInput{
					KeyId:      input.KeyId,
					PolicyName: &policyName,
				})
			}
			return inputs, nil
		},
		GetFunc: getKeyPolicyFunc,
	}
}

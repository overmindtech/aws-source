package networkmanager

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func resourcePolicyGetFunc(ctx context.Context, client *networkmanager.Client, scope, query string) (*string, error) {
	out, err := client.GetResourcePolicy(ctx, &networkmanager.GetResourcePolicyInput{
		ResourceArn: &query,
	})
	if err != nil {
		return nil, err
	}
	return out.PolicyDocument, nil
}

//go:generate docgen ../../docs-data
// +overmind:type networkmanager-resource-policy
// +overmind:descriptiveType Networkmanager Resource Policy
// +overmind:get Get Networkmanager Resource Policy by ARN
// +overmind:group AWS
// +overmind:terraform:queryMap aws_networkmanager_resource_policy.arn

func NewResourcePolicySource(config aws.Config, accountID string, limit *sources.LimitBucket) *sources.GetListSource[*string, *networkmanager.Client, *networkmanager.Options] {
	return &sources.GetListSource[*string, *networkmanager.Client, *networkmanager.Options]{
		Client:    networkmanager.NewFromConfig(config),
		AccountID: accountID,
		ItemType:  "networkmanager-resource-policy",
		GetFunc: func(ctx context.Context, client *networkmanager.Client, scope string, query string) (*string, error) {
			limit.Wait(ctx) // Wait for rate limiting
			return resourcePolicyGetFunc(ctx, client, scope, query)
		},
		ListFunc: func(ctx context.Context, client *networkmanager.Client, scope string) ([]*string, error) {
			return nil, &sdp.QueryError{
				ErrorType:   sdp.QueryError_NOTFOUND,
				ErrorString: "list not supported for  networkmanager-resource-policy, use get",
			}
		},
	}
}

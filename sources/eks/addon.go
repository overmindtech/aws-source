package eks

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func addonGetFunc(ctx context.Context, client EKSClient, scope string, input *eks.DescribeAddonInput) (*sdp.Item, error) {
	out, err := client.DescribeAddon(ctx, input)

	if err != nil {
		return nil, err
	}

	if out.Addon == nil {
		return nil, &sdp.QueryError{
			ErrorType:   sdp.QueryError_NOTFOUND,
			ErrorString: "addon was nil",
		}
	}

	attributes, err := sources.ToAttributesCase(out.Addon)

	if err != nil {
		return nil, err
	}

	// The uniqueAttributeValue for this is a custom field:
	// {clusterName}/{addonName}
	attributes.Set("uniqueName", (*out.Addon.ClusterName + "/" + *out.Addon.AddonName))

	item := sdp.Item{
		Type:            "eks-addon",
		UniqueAttribute: "uniqueName",
		Attributes:      attributes,
		Scope:           scope,
	}

	return &item, nil
}

//go:generate docgen ../../docs-data
// +overmind:type eks-addon
// +overmind:descriptiveType EKS Addon
// +overmind:get Get an addon by unique name ({clusterName}/{addonName})
// +overmind:list List all addons
// +overmind:search Search addons by cluster name
// +overmind:group AWS

func NewAddonSource(config aws.Config, accountID string, region string) *sources.AlwaysGetSource[*eks.ListAddonsInput, *eks.ListAddonsOutput, *eks.DescribeAddonInput, *eks.DescribeAddonOutput, EKSClient, *eks.Options] {
	return &sources.AlwaysGetSource[*eks.ListAddonsInput, *eks.ListAddonsOutput, *eks.DescribeAddonInput, *eks.DescribeAddonOutput, EKSClient, *eks.Options]{
		ItemType:    "eks-addon",
		Client:      eks.NewFromConfig(config),
		AccountID:   accountID,
		Region:      region,
		DisableList: true,
		SearchInputMapper: func(scope, query string) (*eks.ListAddonsInput, error) {
			return &eks.ListAddonsInput{
				ClusterName: &query,
			}, nil
		},
		GetInputMapper: func(scope, query string) *eks.DescribeAddonInput {
			// The uniqueAttributeValue for this is a custom field:
			// {clusterName}/{addonName}
			fields := strings.Split(query, "/")

			var clusterName string
			var addonName string

			if len(fields) == 2 {
				clusterName = fields[0]
				addonName = fields[1]
			}

			return &eks.DescribeAddonInput{
				AddonName:   &addonName,
				ClusterName: &clusterName,
			}
		},
		ListFuncPaginatorBuilder: func(client EKSClient, input *eks.ListAddonsInput) sources.Paginator[*eks.ListAddonsOutput, *eks.Options] {
			return eks.NewListAddonsPaginator(client, input)
		},
		ListFuncOutputMapper: func(output *eks.ListAddonsOutput, input *eks.ListAddonsInput) ([]*eks.DescribeAddonInput, error) {
			inputs := make([]*eks.DescribeAddonInput, len(output.Addons))

			for i := range output.Addons {
				inputs[i] = &eks.DescribeAddonInput{
					AddonName:   &output.Addons[i],
					ClusterName: input.ClusterName,
				}
			}

			return inputs, nil
		},
		GetFunc: addonGetFunc,
	}
}

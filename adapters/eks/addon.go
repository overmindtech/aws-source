package eks

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/overmindtech/aws-source/adapters"
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

	attributes, err := adapters.ToAttributesWithExclude(out.Addon)

	if err != nil {
		return nil, err
	}

	// The uniqueAttributeValue for this is a custom field:
	// {clusterName}/{addonName}
	attributes.Set("UniqueName", (*out.Addon.ClusterName + "/" + *out.Addon.AddonName))

	item := sdp.Item{
		Type:            "eks-addon",
		UniqueAttribute: "UniqueName",
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
// +overmind:terraform:queryMap aws_eks_addon.arn
// +overmind:terraform:method SEARCH

func NewAddonAdapter(client EKSClient, accountID string, region string) *adapters.AlwaysGetAdapter[*eks.ListAddonsInput, *eks.ListAddonsOutput, *eks.DescribeAddonInput, *eks.DescribeAddonOutput, EKSClient, *eks.Options] {
	return &adapters.AlwaysGetAdapter[*eks.ListAddonsInput, *eks.ListAddonsOutput, *eks.DescribeAddonInput, *eks.DescribeAddonOutput, EKSClient, *eks.Options]{
		ItemType:        "eks-addon",
		Client:          client,
		AccountID:       accountID,
		Region:          region,
		AdapterMetadata: AddonMetadata(),
		DisableList:     true,
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
		ListFuncPaginatorBuilder: func(client EKSClient, input *eks.ListAddonsInput) adapters.Paginator[*eks.ListAddonsOutput, *eks.Options] {
			return eks.NewListAddonsPaginator(client, input)
		},
		ListFuncOutputMapper: func(output *eks.ListAddonsOutput, input *eks.ListAddonsInput) ([]*eks.DescribeAddonInput, error) {
			inputs := make([]*eks.DescribeAddonInput, 0, len(output.Addons))

			for i := range output.Addons {
				inputs = append(inputs, &eks.DescribeAddonInput{
					AddonName:   &output.Addons[i],
					ClusterName: input.ClusterName,
				})
			}

			return inputs, nil
		},
		GetFunc: addonGetFunc,
	}
}

func AddonMetadata() sdp.AdapterMetadata {
	return sdp.AdapterMetadata{
		Type:            "eks-addon",
		DescriptiveName: "EKS Addon",
		SupportedQueryMethods: &sdp.AdapterSupportedQueryMethods{
			Get:               true,
			List:              true,
			Search:            true,
			GetDescription:    "Get an addon by unique name ({clusterName}/{addonName})",
			ListDescription:   "List all addons",
			SearchDescription: "Search addons by cluster name",
		},
		TerraformMappings: []*sdp.TerraformMapping{
			{
				TerraformMethod:   sdp.QueryMethod_SEARCH,
				TerraformQueryMap: "aws_eks_addon.arn",
			},
		},
		Category: sdp.AdapterCategory_ADAPTER_CATEGORY_COMPUTE_APPLICATION,
	}
}

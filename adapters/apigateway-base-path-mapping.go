package adapters

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/aws/aws-sdk-go-v2/service/apigateway/types"
	"github.com/overmindtech/aws-source/adapterhelpers"
	"github.com/overmindtech/sdp-go"
)

func convertGetBasePathMappingOutputToBasePathMapping(output *apigateway.GetBasePathMappingOutput) *types.BasePathMapping {
	return &types.BasePathMapping{
		BasePath:  output.BasePath,
		RestApiId: output.RestApiId,
		Stage:     output.Stage,
	}
}

func basePathMappingOutputMapper(query, scope string, awsItem *types.BasePathMapping) (*sdp.Item, error) {
	attributes, err := adapterhelpers.ToAttributesWithExclude(awsItem, "tags")
	if err != nil {
		return nil, err
	}

	domainName := strings.Split(query, "/")[0]

	err = attributes.Set("UniqueAttribute", fmt.Sprintf("%s/%s", domainName, *awsItem.BasePath))

	item := sdp.Item{
		Type:            "apigateway-base-path-mapping",
		UniqueAttribute: "BasePath",
		Attributes:      attributes,
		Scope:           scope,
	}

	item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
		Query: &sdp.Query{
			Type:   "apigateway-domain-name",
			Method: sdp.QueryMethod_GET,
			Query:  domainName,
			Scope:  scope,
		},
		BlastPropagation: &sdp.BlastPropagation{
			// If domain name changes, the base path mapping will be affected
			In: true,
			// If base path mapping changes, the domain name won't be affected
			Out: false,
		},
	})

	if awsItem.RestApiId != nil {
		item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
			Query: &sdp.Query{
				Type:   "apigateway-rest-api",
				Method: sdp.QueryMethod_GET,
				Query:  *awsItem.RestApiId,
				Scope:  scope,
			},
			BlastPropagation: &sdp.BlastPropagation{
				// They are tightly coupled, so we need to propagate the blast to the linked item
				In:  true,
				Out: true,
			},
		})
	}

	return &item, nil
}

func NewAPIGatewayBasePathMappingAdapter(client *apigateway.Client, accountID string, region string) *adapterhelpers.GetListAdapter[*types.BasePathMapping, *apigateway.Client, *apigateway.Options] {
	return &adapterhelpers.GetListAdapter[*types.BasePathMapping, *apigateway.Client, *apigateway.Options]{
		ItemType:        "apigateway-base-path-mapping",
		Client:          client,
		AccountID:       accountID,
		Region:          region,
		AdapterMetadata: basePathMappingAdapterMetadata,
		GetFunc: func(ctx context.Context, client *apigateway.Client, scope, query string) (*types.BasePathMapping, error) {
			f := strings.Split(query, "/")
			if len(f) != 2 {
				return nil, &sdp.QueryError{
					ErrorType:   sdp.QueryError_NOTFOUND,
					ErrorString: fmt.Sprintf("query must be in the format of: the domain-name/base-path, but found: %s", query),
				}
			}
			out, err := client.GetBasePathMapping(ctx, &apigateway.GetBasePathMappingInput{
				DomainName: &f[0],
				BasePath:   &f[1],
			})
			if err != nil {
				return nil, err
			}
			return convertGetBasePathMappingOutputToBasePathMapping(out), nil
		},
		DisableList: true,
		SearchFunc: func(ctx context.Context, client *apigateway.Client, scope string, query string) ([]*types.BasePathMapping, error) {
			out, err := client.GetBasePathMappings(ctx, &apigateway.GetBasePathMappingsInput{
				DomainName: &query,
			})
			if err != nil {
				return nil, err
			}

			var items []*types.BasePathMapping
			for _, basePathMapping := range out.Items {
				items = append(items, &basePathMapping)
			}

			return items, nil
		},
		ItemMapper: func(query, scope string, awsItem *types.BasePathMapping) (*sdp.Item, error) {
			return basePathMappingOutputMapper(query, scope, awsItem)
		},
	}
}

var basePathMappingAdapterMetadata = Metadata.Register(&sdp.AdapterMetadata{
	Type:            "apigateway-base-path-mapping",
	DescriptiveName: "API Gateway Base Path Mapping",
	Category:        sdp.AdapterCategory_ADAPTER_CATEGORY_CONFIGURATION,
	SupportedQueryMethods: &sdp.AdapterSupportedQueryMethods{
		Get:               true,
		Search:            true,
		GetDescription:    "Get an API Gateway Base Path Mapping by its domain name and base path: domain-name/base-path",
		SearchDescription: "Search for API Gateway Base Path Mappings by their domain name: domain-name",
	},
})

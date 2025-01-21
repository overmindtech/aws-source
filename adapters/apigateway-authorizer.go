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

// convertGetAuthorizerOutputToAuthorizer converts a GetAuthorizerOutput to an Authorizer
func convertGetAuthorizerOutputToAuthorizer(output *apigateway.GetAuthorizerOutput) *types.Authorizer {
	return &types.Authorizer{
		Id:                           output.Id,
		Name:                         output.Name,
		Type:                         output.Type,
		ProviderARNs:                 output.ProviderARNs,
		AuthType:                     output.AuthType,
		AuthorizerUri:                output.AuthorizerUri,
		AuthorizerCredentials:        output.AuthorizerCredentials,
		IdentitySource:               output.IdentitySource,
		IdentityValidationExpression: output.IdentityValidationExpression,
		AuthorizerResultTtlInSeconds: output.AuthorizerResultTtlInSeconds,
	}
}

func authorizerOutputMapper(scope string, awsItem *types.Authorizer) (*sdp.Item, error) {
	attributes, err := adapterhelpers.ToAttributesWithExclude(awsItem, "tags")
	if err != nil {
		return nil, err
	}

	item := sdp.Item{
		Type:            "apigateway-authorizer",
		UniqueAttribute: "Id",
		Attributes:      attributes,
		Scope:           scope,
	}

	return &item, nil
}

func NewAPIGatewayAuthorizerAdapter(client *apigateway.Client, accountID string, region string) *adapterhelpers.GetListAdapter[*types.Authorizer, *apigateway.Client, *apigateway.Options] {
	return &adapterhelpers.GetListAdapter[*types.Authorizer, *apigateway.Client, *apigateway.Options]{
		ItemType:        "apigateway-authorizer",
		Client:          client,
		AccountID:       accountID,
		Region:          region,
		AdapterMetadata: authorizerAdapterMetadata,
		GetFunc: func(ctx context.Context, client *apigateway.Client, scope, query string) (*types.Authorizer, error) {
			f := strings.Split(query, "/")
			if len(f) != 2 {
				return nil, &sdp.QueryError{
					ErrorType:   sdp.QueryError_NOTFOUND,
					ErrorString: fmt.Sprintf("query must be in the format of: the rest-api-id/authorizer-id, but found: %s", query),
				}
			}
			out, err := client.GetAuthorizer(ctx, &apigateway.GetAuthorizerInput{
				RestApiId:    &f[0],
				AuthorizerId: &f[1],
			})
			if err != nil {
				return nil, err
			}
			return convertGetAuthorizerOutputToAuthorizer(out), nil
		},
		DisableList: true,
		SearchFunc: func(ctx context.Context, client *apigateway.Client, scope string, query string) ([]*types.Authorizer, error) {
			f := strings.Split(query, "/")
			var restAPIID string
			var name string

			switch len(f) {
			case 1:
				restAPIID = f[0]
			case 2:
				restAPIID = f[0]
				name = f[1]
			default:
				return nil, &sdp.QueryError{
					ErrorType: sdp.QueryError_NOTFOUND,
					ErrorString: fmt.Sprintf(
						"query must be in the format of: the rest-api-id/authorizer-id or rest-api-id, but found: %s",
						query,
					),
				}
			}

			out, err := client.GetAuthorizers(ctx, &apigateway.GetAuthorizersInput{
				RestApiId: &restAPIID,
			})
			if err != nil {
				return nil, err
			}

			var items []*types.Authorizer
			for _, authorizer := range out.Items {
				if name != "" && strings.Contains(*authorizer.Name, name) {
					items = append(items, &authorizer)
				} else {
					items = append(items, &authorizer)
				}
			}

			return items, nil
		},
		ItemMapper: func(_, scope string, awsItem *types.Authorizer) (*sdp.Item, error) {
			return authorizerOutputMapper(scope, awsItem)
		},
	}
}

var authorizerAdapterMetadata = Metadata.Register(&sdp.AdapterMetadata{
	Type:            "apigateway-authorizer",
	DescriptiveName: "API Gateway Authorizer",
	Category:        sdp.AdapterCategory_ADAPTER_CATEGORY_SECURITY,
	SupportedQueryMethods: &sdp.AdapterSupportedQueryMethods{
		Get:               true,
		Search:            true,
		GetDescription:    "Get an API Gateway Authorizer by its rest API ID and ID: rest-api-id/authorizer-id",
		SearchDescription: "Search for API Gateway Authorizers by their rest API ID or with rest API ID and their name: rest-api-id/authorizer-name",
	},
	TerraformMappings: []*sdp.TerraformMapping{
		{TerraformQueryMap: "aws_api_gateway_authorizer.id"},
	},
})

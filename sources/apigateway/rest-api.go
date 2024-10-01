package apigateway

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/aws/aws-sdk-go-v2/service/apigateway/types"
	"github.com/micahhausler/aws-iam-policy/policy"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/aws-source/sources/iam"
	"github.com/overmindtech/sdp-go"

	log "github.com/sirupsen/logrus"
)

// convertGetRestApiOutputToRestApi converts a GetRestApiOutput to a RestApi
func convertGetRestApiOutputToRestApi(output *apigateway.GetRestApiOutput) *types.RestApi {
	return &types.RestApi{
		CreatedDate:               output.CreatedDate,
		Description:               output.Description,
		Id:                        output.Id,
		Name:                      output.Name,
		Tags:                      output.Tags,
		ApiKeySource:              output.ApiKeySource,
		BinaryMediaTypes:          output.BinaryMediaTypes,
		DisableExecuteApiEndpoint: output.DisableExecuteApiEndpoint,
		EndpointConfiguration:     output.EndpointConfiguration,
		MinimumCompressionSize:    output.MinimumCompressionSize,
		Policy:                    output.Policy,
		RootResourceId:            output.RootResourceId,
		Version:                   output.Version,
		Warnings:                  output.Warnings,
	}
}

func restApiListFunc(ctx context.Context, client *apigateway.Client, _ string) ([]*types.RestApi, error) {
	out, err := client.GetRestApis(ctx, &apigateway.GetRestApisInput{})
	if err != nil {
		return nil, err
	}

	var items []*types.RestApi
	for _, restAPI := range out.Items {
		items = append(items, &restAPI)
	}

	return items, nil
}

func restApiOutputMapper(scope string, awsItem *types.RestApi) (*sdp.Item, error) {
	attributes, err := sources.ToAttributesWithExclude(awsItem, "tags")
	if err != nil {
		return nil, err
	}

	if awsItem.Policy != nil {
		type restAPIWithParsedPolicy struct {
			*types.RestApi
			PolicyDocument *policy.Policy
		}

		restApi := restAPIWithParsedPolicy{
			RestApi: awsItem,
		}

		restApi.PolicyDocument, err = iam.ParsePolicyDocument(*awsItem.Policy)
		if err != nil {
			log.WithFields(log.Fields{
				"error":          err,
				"scope":          scope,
				"policyDocument": *awsItem.Policy,
			}).Error("Error parsing policy document")

			return nil, nil //nolint:nilerr
		}

		attributes, err = sources.ToAttributesWithExclude(restApi, "tags")
		if err != nil {
			return nil, err
		}
	}

	item := sdp.Item{
		Type:            "apigateway-rest-api",
		UniqueAttribute: "Id",
		Attributes:      attributes,
		Scope:           scope,
		Tags:            awsItem.Tags,
	}

	if awsItem.EndpointConfiguration != nil && awsItem.EndpointConfiguration.VpcEndpointIds != nil {
		for _, vpcEndpointID := range awsItem.EndpointConfiguration.VpcEndpointIds {
			// +overmind:link ec2-vpc-endpoint
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "ec2-vpc-endpoint",
					Method: sdp.QueryMethod_GET,
					Query:  vpcEndpointID,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					// Any change on the VPC endpoint should affect the REST API
					In: true,
					// We can't affect the VPC endpoint
					Out: false,
				},
			})
		}
	}

	if awsItem.RootResourceId != nil {
		// +overmind:link apigateway-resource
		item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
			Query: &sdp.Query{
				Type:   "apigateway-resource",
				Method: sdp.QueryMethod_GET,
				Query:  fmt.Sprintf("%s/%s", *awsItem.Id, *awsItem.RootResourceId),
				Scope:  scope,
			},
			BlastPropagation: &sdp.BlastPropagation{
				// They are tightly linked
				In:  true,
				Out: true,
			},
		})
	}

	// +overmind:link apigateway-resource
	item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
		Query: &sdp.Query{
			Type:   "apigateway-resource",
			Method: sdp.QueryMethod_SEARCH,
			Query:  *awsItem.Id,
			Scope:  scope,
		},
		BlastPropagation: &sdp.BlastPropagation{
			// Updating a resource won't affect the REST API
			In: false,
			// Updating the REST API will affect the resources
			Out: true,
		},
	})

	return &item, nil
}

//go:generate docgen ../../docs-data
// +overmind:type apigateway-rest-api
// +overmind:descriptiveType REST API
// +overmind:get Get a REST API by ID
// +overmind:list List all REST APIs
// +overmind:search Search for REST APIs their name
// +overmind:group AWS
// +overmind:terraform:queryMap aws_api_gateway_rest_api.id

func NewRestApiSource(client *apigateway.Client, accountID string, region string) *sources.GetListSource[*types.RestApi, *apigateway.Client, *apigateway.Options] {
	return &sources.GetListSource[*types.RestApi, *apigateway.Client, *apigateway.Options]{
		ItemType:  "apigateway-rest-api",
		Client:    client,
		AccountID: accountID,
		Region:    region,
		GetFunc: func(ctx context.Context, client *apigateway.Client, scope, query string) (*types.RestApi, error) {
			out, err := client.GetRestApi(ctx, &apigateway.GetRestApiInput{
				RestApiId: &query,
			})
			if err != nil {
				return nil, err
			}
			return convertGetRestApiOutputToRestApi(out), nil
		},
		ListFunc: restApiListFunc,
		SearchFunc: func(ctx context.Context, client *apigateway.Client, scope string, query string) ([]*types.RestApi, error) {
			out, err := client.GetRestApis(ctx, &apigateway.GetRestApisInput{})
			if err != nil {
				return nil, err
			}

			var items []*types.RestApi
			for _, restAPI := range out.Items {
				if *restAPI.Name == query {
					items = append(items, &restAPI)
				}
			}

			return items, nil
		},
		ItemMapper: func(_, scope string, awsItem *types.RestApi) (*sdp.Item, error) {
			return restApiOutputMapper(scope, awsItem)
		},
	}
}

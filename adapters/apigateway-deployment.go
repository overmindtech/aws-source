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

// convertGetDeploymentOutputToDeployment converts a GetDeploymentOutput to a Deployment
func convertGetDeploymentOutputToDeployment(output *apigateway.GetDeploymentOutput) *types.Deployment {
	return &types.Deployment{
		Id:          output.Id,
		CreatedDate: output.CreatedDate,
		Description: output.Description,
		ApiSummary: output.ApiSummary,
	}
}

func deploymentOutputMapper(scope string, awsItem *types.Deployment) (*sdp.Item, error) {
	attributes, err := adapterhelpers.ToAttributesWithExclude(awsItem, "tags")
	if err != nil {
		return nil, err
	}

	item := sdp.Item{
		Type:            "apigateway-deployment",
		UniqueAttribute: "Id",
		Attributes:      attributes,
		Scope:           scope,
	}

	return &item, nil
}

func NewAPIGatewayDeploymentAdapter(client *apigateway.Client, accountID string, region string) *adapterhelpers.GetListAdapter[*types.Deployment, *apigateway.Client, *apigateway.Options] {
	return &adapterhelpers.GetListAdapter[*types.Deployment, *apigateway.Client, *apigateway.Options]{
		ItemType:        "apigateway-deployment",
		Client:          client,
		AccountID:       accountID,
		Region:          region,
		AdapterMetadata: deploymentAdapterMetadata,
		GetFunc: func(ctx context.Context, client *apigateway.Client, scope, query string) (*types.Deployment, error) {
			f := strings.Split(query, "/")
			if len(f) != 2 {
				return nil, &sdp.QueryError{
					ErrorType:   sdp.QueryError_NOTFOUND,
					ErrorString: fmt.Sprintf("query must be in the format of: the rest-api-id/deployment-id, but found: %s", query),
				}
			}
			out, err := client.GetDeployment(ctx, &apigateway.GetDeploymentInput{
				RestApiId:    &f[0],
				DeploymentId: &f[1],
			})
			if err != nil {
				return nil, err
			}
			return convertGetDeploymentOutputToDeployment(out), nil
		},
		DisableList: true,
		SearchFunc: func(ctx context.Context, client *apigateway.Client, scope string, query string) ([]*types.Deployment, error) {
			f := strings.Split(query, "/")
			var restAPIID string
			var description string

			switch len(f) {
			case 1:
				restAPIID = f[0]
			case 2:
				restAPIID = f[0]
				description = f[1]
			default:
				return nil, &sdp.QueryError{
					ErrorType: sdp.QueryError_NOTFOUND,
					ErrorString: fmt.Sprintf(
						"query must be in the format of: the rest-api-id/deployment-id or rest-api-id, but found: %s",
						query,
					),
				}
			}

			out, err := client.GetDeployments(ctx, &apigateway.GetDeploymentsInput{
				RestApiId: &restAPIID,
			})
			if err != nil {
				return nil, err
			}

			var items []*types.Deployment
			for _, deployment := range out.Items {
				if description != "" && strings.Contains(*deployment.Description, description) {
					items = append(items, &deployment)
				} else {
					items = append(items, &deployment)
				}
			}

			return items, nil
		},
		ItemMapper: func(_, scope string, awsItem *types.Deployment) (*sdp.Item, error) {
			return deploymentOutputMapper(scope, awsItem)
		},
	}
}

var deploymentAdapterMetadata = Metadata.Register(&sdp.AdapterMetadata{
	Type:            "apigateway-deployment",
	DescriptiveName: "API Gateway Deployment",
	Category:        sdp.AdapterCategory_ADAPTER_CATEGORY_CONFIGURATION,
	SupportedQueryMethods: &sdp.AdapterSupportedQueryMethods{
		Get:               true,
		Search:            true,
		GetDescription:    "Get an API Gateway Deployment by its rest API ID and ID: rest-api-id/deployment-id",
		SearchDescription: "Search for API Gateway Deployments by their rest API ID or with rest API ID and their description: rest-api-id/deployment-description",
	},
})

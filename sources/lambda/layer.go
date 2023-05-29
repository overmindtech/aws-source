package lambda

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func layerListFunc(ctx context.Context, client *lambda.Client, scope string) ([]*types.LayersListItem, error) {
	paginator := lambda.NewListLayersPaginator(client, &lambda.ListLayersInput{})
	layers := make([]*types.LayersListItem, 0)

	for paginator.HasMorePages() {
		out, err := paginator.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, layer := range out.Layers {
			layers = append(layers, &layer)
		}
	}

	return layers, nil
}

func layerItemMapper(scope string, awsItem *types.LayersListItem) (*sdp.Item, error) {
	attributes, err := sources.ToAttributesCase(awsItem)

	if err != nil {
		return nil, err
	}

	item := sdp.Item{
		Type:            "lambda-layer",
		UniqueAttribute: "layerName",
		Attributes:      attributes,
		Scope:           scope,
	}

	if awsItem.LatestMatchingVersion != nil {
		// +overmind:link lambda-layer-version
		item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
			Query: &sdp.Query{
				Type:   "lambda-layer-version",
				Method: sdp.QueryMethod_GET,
				Query:  fmt.Sprintf("%v:%v", *awsItem.LayerName, awsItem.LatestMatchingVersion.Version),
				Scope:  scope,
			},
			BlastPropagation: &sdp.BlastPropagation{
				// Tightly coupled
				In:  true,
				Out: true,
			},
		})
	}

	return &item, nil
}

//go:generate docgen ../../docs-data
// +overmind:type lambda-layer
// +overmind:descriptiveType Lambda Layer
// +overmind:list List all lambda layers
// +overmind:group AWS

func NewLayerSource(config aws.Config, accountID string, region string) *sources.GetListSource[*types.LayersListItem, *lambda.Client, *lambda.Options] {
	return &sources.GetListSource[*types.LayersListItem, *lambda.Client, *lambda.Options]{
		ItemType:  "lambda-layer",
		Client:    lambda.NewFromConfig(config),
		AccountID: accountID,
		Region:    region,
		GetFunc: func(_ context.Context, _ *lambda.Client, _, _ string) (*types.LayersListItem, error) {
			// Layers can only be listed
			return nil, errors.New("get is not supported for lambda-layers")
		},
		ListFunc:   layerListFunc,
		ItemMapper: layerItemMapper,
	}
}

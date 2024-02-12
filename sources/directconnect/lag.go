package directconnect

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/directconnect/types"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/directconnect"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func lagOutputMapper(_ context.Context, _ *directconnect.Client, scope string, _ *directconnect.DescribeLagsInput, output *directconnect.DescribeLagsOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, lag := range output.Lags {
		attributes, err := sources.ToAttributesCase(lag, "tags")
		if err != nil {
			return nil, err
		}

		item := sdp.Item{
			Type:            "directconnect-lag",
			UniqueAttribute: "lagId",
			Attributes:      attributes,
			Scope:           scope,
			Tags:            tagsToMap(lag.Tags),
		}

		switch lag.LagState {
		case types.LagStateRequested:
			item.Health = sdp.Health_HEALTH_PENDING.Enum()
		case types.LagStatePending:
			item.Health = sdp.Health_HEALTH_PENDING.Enum()
		case types.LagStateAvailable:
			item.Health = sdp.Health_HEALTH_OK.Enum()
		case types.LagStateDown:
			item.Health = sdp.Health_HEALTH_ERROR.Enum()
		case types.LagStateDeleting:
			item.Health = sdp.Health_HEALTH_UNKNOWN.Enum()
		case types.LagStateDeleted:
			item.Health = sdp.Health_HEALTH_UNKNOWN.Enum()
		case types.LagStateUnknown:
			item.Health = sdp.Health_HEALTH_UNKNOWN.Enum()
		}

		for _, connection := range lag.Connections {
			if connection.ConnectionId != nil {
				// +overmind:link directconnect-connection
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "directconnect-connection",
						Method: sdp.QueryMethod_GET,
						Query:  *connection.ConnectionId,
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						// Connection and LAG are tightly coupled
						// Changing one will affect the other
						In:  true,
						Out: true,
					},
				})
			}
		}

		if lag.LagId != nil {
			// +overmind:link directconnect-hosted-connection
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "directconnect-hosted-connection",
					Method: sdp.QueryMethod_SEARCH,
					Query:  *lag.LagId,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					// LAG and hosted connections are tightly coupled
					// Changing one will affect the other
					In:  true,
					Out: true,
				},
			})
		}

		if lag.Location != nil {
			// +overmind:link directconnect-location
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "directconnect-location",
					Method: sdp.QueryMethod_GET,
					// This is location code, not its name
					Query: *lag.Location,
					Scope: scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					// Changes to the location will affect this, i.e., its speed, provider, etc.
					In: true,
					// We can't affect the location
					Out: false,
				},
			})
		}

		items = append(items, &item)
	}

	return items, nil
}

//go:generate docgen ../../docs-data
// +overmind:type directconnect-lag
// +overmind:descriptiveType Direct Connect Link Aggregation Group
// +overmind:get Get a Link Aggregation Group by ID
// +overmind:list List all Link Aggregation Groups
// +overmind:search Search Link Aggregation Group by ARN
// +overmind:group AWS
// +overmind:terraform:queryMap aws_dx_lag.id

func NewLagSource(config aws.Config, accountID string, limit *sources.LimitBucket) *sources.DescribeOnlySource[*directconnect.DescribeLagsInput, *directconnect.DescribeLagsOutput, *directconnect.Client, *directconnect.Options] {
	return &sources.DescribeOnlySource[*directconnect.DescribeLagsInput, *directconnect.DescribeLagsOutput, *directconnect.Client, *directconnect.Options]{
		Config:    config,
		Client:    directconnect.NewFromConfig(config),
		AccountID: accountID,
		ItemType:  "directconnect-lag",
		DescribeFunc: func(ctx context.Context, client *directconnect.Client, input *directconnect.DescribeLagsInput) (*directconnect.DescribeLagsOutput, error) {
			limit.Wait(ctx) // Wait for rate limiting
			return client.DescribeLags(ctx, input)
		},
		InputMapperGet: func(scope, query string) (*directconnect.DescribeLagsInput, error) {
			return &directconnect.DescribeLagsInput{
				LagId: &query,
			}, nil
		},
		InputMapperList: func(scope string) (*directconnect.DescribeLagsInput, error) {
			return &directconnect.DescribeLagsInput{}, nil
		},
		OutputMapper: lagOutputMapper,
	}
}

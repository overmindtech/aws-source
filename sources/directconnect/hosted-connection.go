package directconnect

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/directconnect"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func hostedConnectionOutputMapper(_ context.Context, _ *directconnect.Client, scope string, _ *directconnect.DescribeHostedConnectionsInput, output *directconnect.DescribeHostedConnectionsOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, connection := range output.Connections {
		attributes, err := sources.ToAttributesCase(connection, "tags")
		if err != nil {
			return nil, err
		}

		item := sdp.Item{
			Type:            "directconnect-hosted-connection",
			UniqueAttribute: "connectionId",
			Attributes:      attributes,
			Scope:           scope,
			Tags:            tagsToMap(connection.Tags),
		}

		if connection.LagId != nil {
			// +overmind:link directconnect-lag
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "directconnect-lag",
					Method: sdp.QueryMethod_GET,
					Query:  *connection.LagId,
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

		if connection.Location != nil {
			// +overmind:link directconnect-location
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "directconnect-location",
					Method: sdp.QueryMethod_GET,
					Query:  *connection.Location,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					// Changes to the location will affect this, i.e., its speed, provider, etc.
					In: true,
					// We can't affect the location
					Out: false,
				},
			})
		}

		if connection.LoaIssueTime != nil {
			// +overmind:link directconnect-loa
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "directconnect-loa",
					Method: sdp.QueryMethod_GET,
					Query:  *connection.ConnectionId,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					// Changes to the loa will affect this
					In: true,
					// We can't affect the loa
					Out: false,
				},
			})
		}

		item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
			// +overmind:link directconnect-virtual-interface
			Query: &sdp.Query{
				Type:   "directconnect-virtual-interface",
				Method: sdp.QueryMethod_SEARCH,
				Query:  *connection.ConnectionId,
				Scope:  scope,
			},
			BlastPropagation: &sdp.BlastPropagation{
				// Changes to the virtual interface won't affect this
				In: false,
				// We cannot delete a hosted connection if it has virtual interfaces
				Out: true,
			},
		})

		items = append(items, &item)
	}

	return items, nil
}

//go:generate docgen ../../docs-data
// +overmind:type directconnect-hosted-connection
// +overmind:descriptiveType Direct Connect Hosted Connection
// +overmind:get Get a Hosted Connection by connection ID
// +overmind:list List all
// +overmind:search Search Hosted Connections by ARN
// +overmind:group AWS
// +overmind:terraform:queryMap aws_dx_hosted_connection.id

func NewHostedConnectionSource(config aws.Config, accountID string, limit *sources.LimitBucket) *sources.DescribeOnlySource[*directconnect.DescribeHostedConnectionsInput, *directconnect.DescribeHostedConnectionsOutput, *directconnect.Client, *directconnect.Options] {
	return &sources.DescribeOnlySource[*directconnect.DescribeHostedConnectionsInput, *directconnect.DescribeHostedConnectionsOutput, *directconnect.Client, *directconnect.Options]{
		Config:    config,
		Client:    directconnect.NewFromConfig(config),
		AccountID: accountID,
		ItemType:  "directconnect-hosted-connection",
		DescribeFunc: func(ctx context.Context, client *directconnect.Client, input *directconnect.DescribeHostedConnectionsInput) (*directconnect.DescribeHostedConnectionsOutput, error) {
			limit.Wait(ctx) // Wait for rate limiting
			return client.DescribeHostedConnections(ctx, input)
		},
		InputMapperGet: func(scope, query string) (*directconnect.DescribeHostedConnectionsInput, error) {
			return &directconnect.DescribeHostedConnectionsInput{
				ConnectionId: &query,
			}, nil
		},
		InputMapperList: func(scope string) (*directconnect.DescribeHostedConnectionsInput, error) {
			return &directconnect.DescribeHostedConnectionsInput{}, nil
		},
		OutputMapper: hostedConnectionOutputMapper,
	}
}

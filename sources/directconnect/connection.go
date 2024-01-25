package directconnect

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/directconnect"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func connectionOutputMapper(_ context.Context, _ *directconnect.Client, scope string, _ *directconnect.DescribeConnectionsInput, output *directconnect.DescribeConnectionsOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, connection := range output.Connections {
		attributes, err := sources.ToAttributesCase(connection, "tags")
		if err != nil {
			return nil, err
		}

		item := sdp.Item{
			Type:            "directconnect-connection",
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
					// Changes to the lag will affect this
					In: true,
					// We can't affect the lag
					Out: false,
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

		// Virtual Interfaces
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
				// We cannot delete a connection if it has virtual interfaces
				Out: true,
			},
		})

		items = append(items, &item)
	}

	return items, nil
}

//go:generate docgen ../../docs-data
// +overmind:type directconnect-connection
// +overmind:descriptiveType Connection
// +overmind:get Get a connection by ID
// +overmind:list List all connections
// +overmind:search Search connection by ARN
// +overmind:group AWS
// +overmind:terraform:queryMap aws_dx_connection.id

func NewConnectionSource(config aws.Config, accountID string, limit *sources.LimitBucket) *sources.DescribeOnlySource[*directconnect.DescribeConnectionsInput, *directconnect.DescribeConnectionsOutput, *directconnect.Client, *directconnect.Options] {
	return &sources.DescribeOnlySource[*directconnect.DescribeConnectionsInput, *directconnect.DescribeConnectionsOutput, *directconnect.Client, *directconnect.Options]{
		Config:    config,
		Client:    directconnect.NewFromConfig(config),
		AccountID: accountID,
		ItemType:  "directconnect-connection",
		DescribeFunc: func(ctx context.Context, client *directconnect.Client, input *directconnect.DescribeConnectionsInput) (*directconnect.DescribeConnectionsOutput, error) {
			limit.Wait(ctx) // Wait for rate limiting
			return client.DescribeConnections(ctx, input)
		},
		InputMapperGet: func(scope, query string) (*directconnect.DescribeConnectionsInput, error) {
			return &directconnect.DescribeConnectionsInput{
				ConnectionId: &query,
			}, nil
		},
		InputMapperList: func(scope string) (*directconnect.DescribeConnectionsInput, error) {
			return &directconnect.DescribeConnectionsInput{}, nil
		},
		OutputMapper: connectionOutputMapper,
	}
}

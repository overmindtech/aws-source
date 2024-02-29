package networkmanager

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func siteOutputMapper(_ context.Context, _ NetworkmanagerClient, scope string, _ *networkmanager.GetSitesInput, output *networkmanager.GetSitesOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, s := range output.Sites {
		var err error
		var attrs *sdp.ItemAttributes
		attrs, err = sources.ToAttributesCase(s, "tags")

		if err != nil {
			return nil, &sdp.QueryError{
				ErrorType:   sdp.QueryError_OTHER,
				ErrorString: err.Error(),
				Scope:       scope,
			}
		}

		attrs.Set("globalNetworkSiteId", fmt.Sprintf("%s/%s", *s.GlobalNetworkId, *s.SiteId))

		item := sdp.Item{
			Type:            "networkmanager-site",
			UniqueAttribute: "globalNetworkSiteId",
			Scope:           scope,
			Attributes:      attrs,
			Tags:            tagsToMap(s.Tags),
			LinkedItemQueries: []*sdp.LinkedItemQuery{
				{
					Query: &sdp.Query{
						// +overmind:link networkmanager-global-network
						// Search for all sites with this global network
						Type:   "networkmanager-global-network",
						Method: sdp.QueryMethod_GET,
						Query:  *s.GlobalNetworkId,
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						// ?? Sites can affect the global network
						In: true,
						// The global network will definitely affect the site
						// instances
						Out: true,
					},
				},
			},
		}
		switch s.State {
		case types.SiteStatePending:
			item.Health = sdp.Health_HEALTH_PENDING.Enum()
		case types.SiteStateAvailable:
			item.Health = sdp.Health_HEALTH_OK.Enum()
		case types.SiteStateUpdating:
			item.Health = sdp.Health_HEALTH_PENDING.Enum()
		case types.SiteStateDeleting:
			item.Health = sdp.Health_HEALTH_PENDING.Enum()
		}

		items = append(items, &item)
	}

	return items, nil
}

//go:generate docgen ../../docs-data
// +overmind:type networkmanager-site
// +overmind:descriptiveType Networkmanager Site
// +overmind:get Get a Networkmanager Site
// +overmind:list List all Networkmanager Sites
// +overmind:search Search for Networkmanager Sites by GlobalNetworkId
// +overmind:group AWS

func NewSiteSource(config aws.Config, accountID string, limit *sources.LimitBucket) *sources.DescribeOnlySource[*networkmanager.GetSitesInput, *networkmanager.GetSitesOutput, NetworkmanagerClient, *networkmanager.Options] {
	return &sources.DescribeOnlySource[*networkmanager.GetSitesInput, *networkmanager.GetSitesOutput, NetworkmanagerClient, *networkmanager.Options]{
		Client:    networkmanager.NewFromConfig(config),
		AccountID: accountID,
		ItemType:  "networkmanager-sites",
		DescribeFunc: func(ctx context.Context, client NetworkmanagerClient, input *networkmanager.GetSitesInput) (*networkmanager.GetSitesOutput, error) {
			return client.GetSites(ctx, input)
		},
		InputMapperGet: func(scope, query string) (*networkmanager.GetSitesInput, error) {
			// We are using a custom id of {globalNetworkId}/{siteId} e.g.
			sections := strings.Split(query, "/")

			if len(sections) != 2 {
				return nil, &sdp.QueryError{
					ErrorType:   sdp.QueryError_NOTFOUND,
					ErrorString: "invalid query for networkmanager-site get function",
				}
			}
			return &networkmanager.GetSitesInput{
				GlobalNetworkId: &sections[0],
				SiteIds: []string{
					sections[1],
				},
			}, nil
		},
		InputMapperList: func(scope string) (*networkmanager.GetSitesInput, error) {
			return nil, &sdp.QueryError{
				ErrorType:   sdp.QueryError_NOTFOUND,
				ErrorString: "list not supported for networkmanager-site, use search",
			}
		},
		PaginatorBuilder: func(client NetworkmanagerClient, params *networkmanager.GetSitesInput) sources.Paginator[*networkmanager.GetSitesOutput, *networkmanager.Options] {
			return networkmanager.NewGetSitesPaginator(client, params)
		},
		OutputMapper: siteOutputMapper,
		InputMapperSearch: func(ctx context.Context, client NetworkmanagerClient, scope, query string) (*networkmanager.GetSitesInput, error) {
			// Search by GlobalNetworkId
			return &networkmanager.GetSitesInput{
				GlobalNetworkId: &query,
			}, nil
		},
	}
}

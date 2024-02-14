package networkmanager

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
	"strings"
)

func sitesGetFunc(ctx context.Context, client NetworkmanagerClient, scope string, input *networkmanager.GetSitesInput) (*sdp.Item, error) {
	if input == nil || input.GlobalNetworkId == nil {
		return nil, &sdp.QueryError{
			ErrorType:   sdp.QueryError_NOTFOUND,
			ErrorString: "GetSitesInput.GlobalNetworkId param is mandatory",
		}
	}
	out, err := client.GetSites(ctx, input)
	if err != nil {
		return nil, err
	}
	if len(out.Sites) == 0 {
		return nil, &sdp.QueryError{
			ErrorType:   sdp.QueryError_NOTFOUND,
			ErrorString: fmt.Sprintf("site with ID %s not found", input.SiteIds[0]),
		}
	}
	if len(out.Sites) != 1 {
		return nil, fmt.Errorf("got %d sites, expected 1", len(out.Sites))
	}

	site := out.Sites[0]

	attributes, err := sources.ToAttributesCase(site, "tags")
	if err != nil {
		return nil, err
	}

	item := sdp.Item{
		Type:            "networkmanager-site",
		UniqueAttribute: "siteId",
		Scope:           scope,
		Attributes:      attributes,
		Tags:            tagsToMap(site.Tags),
		LinkedItemQueries: []*sdp.LinkedItemQuery{
			{
				Query: &sdp.Query{
					// +overmind:link networkmanager-global-network
					// Search for all sites with this global network
					Type:   "networkmanager-global-network",
					Method: sdp.QueryMethod_GET,
					Query:  *site.GlobalNetworkId,
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

	switch site.State {
	case types.SiteStatePending:
		item.Health = sdp.Health_HEALTH_PENDING.Enum()
	case types.SiteStateAvailable:
		item.Health = sdp.Health_HEALTH_OK.Enum()
	case types.SiteStateUpdating:
		item.Health = sdp.Health_HEALTH_PENDING.Enum()
	case types.SiteStateDeleting:
		item.Health = sdp.Health_HEALTH_PENDING.Enum()
	}

	return &item, nil
}

//go:generate docgen ../../docs-data
// +overmind:type networkmanager-site
// +overmind:descriptiveType Networkmanager Site
// +overmind:get Get a Networkmanager Site
// +overmind:list List a Networkmanager Sites
// +overmind:group AWS
// +overmind:terraform:queryMap aws_networkmanager_site.id

func NewSiteSource(config aws.Config, accountID string, limit *sources.LimitBucket) *sources.AlwaysGetSource[*networkmanager.GetSitesInput, *networkmanager.GetSitesOutput, *networkmanager.GetSitesInput, *networkmanager.GetSitesOutput, NetworkmanagerClient, *networkmanager.Options] {
	return &sources.AlwaysGetSource[*networkmanager.GetSitesInput, *networkmanager.GetSitesOutput, *networkmanager.GetSitesInput, *networkmanager.GetSitesOutput, NetworkmanagerClient, *networkmanager.Options]{
		Client:    networkmanager.NewFromConfig(config),
		AccountID: accountID,
		ItemType:  "networkmanager-sites",
		GetFunc:   sitesGetFunc,
		GetInputMapper: func(scope, query string) *networkmanager.GetSitesInput {
			// We are using a custom id of {globalNetworkId}/{siteId} e.g.
			sections := strings.Split(query, "/")

			if len(sections) != 2 {
				return nil
			}
			return &networkmanager.GetSitesInput{
				GlobalNetworkId: &sections[0],
				SiteIds: []string{
					sections[1],
				},
			}
		},
		ListInput: &networkmanager.GetSitesInput{},
		ListFuncPaginatorBuilder: func(client NetworkmanagerClient, input *networkmanager.GetSitesInput) sources.Paginator[*networkmanager.GetSitesOutput, *networkmanager.Options] {
			// GlobalNetworkId is required
			// SiteIds must be valid slice, not nil
			if input.GlobalNetworkId == nil {
				input.GlobalNetworkId = aws.String("")
				input.SiteIds = []string{}
			}
			return networkmanager.NewGetSitesPaginator(client, input)
		},
		ListFuncOutputMapper: func(output *networkmanager.GetSitesOutput, input *networkmanager.GetSitesInput) ([]*networkmanager.GetSitesInput, error) {
			inputs := make([]*networkmanager.GetSitesInput, 0)
			for _, s := range output.Sites {
				inputs = append(inputs, &networkmanager.GetSitesInput{
					SiteIds: []string{
						*s.SiteId, // This will be the id of the site
					},
				})
			}

			return inputs, nil
		},
		SearchInputMapper: func(scope, query string) (*networkmanager.GetSitesInput, error) {
			return &networkmanager.GetSitesInput{
				GlobalNetworkId: &query,
			}, nil
		},
	}
}

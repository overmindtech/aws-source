package networkmanager

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func linkOutputMapper(_ context.Context, _ *networkmanager.Client, scope string, _ *networkmanager.GetLinksInput, output *networkmanager.GetLinksOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, s := range output.Links {
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

		attrs.Set("globalNetworkIdLinkId", idWithGlobalNetwork(*s.GlobalNetworkId, *s.LinkId))

		item := sdp.Item{
			Type:            "networkmanager-link",
			UniqueAttribute: "globalNetworkIdLinkId",
			Scope:           scope,
			Attributes:      attrs,
			Tags:            tagsToMap(s.Tags),
			LinkedItemQueries: []*sdp.LinkedItemQuery{
				{
					Query: &sdp.Query{
						// +overmind:link networkmanager-global-network
						Type:   "networkmanager-global-network",
						Method: sdp.QueryMethod_GET,
						Query:  *s.GlobalNetworkId,
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						In:  true,
						Out: false,
					},
				},
				{
					Query: &sdp.Query{
						// +overmind:link networkmanager-site
						Type:   "networkmanager-site",
						Method: sdp.QueryMethod_GET,
						Query:  idWithGlobalNetwork(*s.GlobalNetworkId, *s.SiteId),
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						In:  true,
						Out: true,
					},
				},
				{
					Query: &sdp.Query{
						// +overmind:link networkmanager-network-resource-relationship
						Type:   "networkmanager-network-resource-relationship",
						Method: sdp.QueryMethod_GET,
						Query:  idWithGlobalNetwork(*s.GlobalNetworkId, *s.LinkArn),
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						In:  true,
						Out: true,
					},
				},
				{
					Query: &sdp.Query{
						// +overmind:link networkmanager-link-association
						Type:   "networkmanager-link-association",
						Method: sdp.QueryMethod_SEARCH,
						Query:  idWithTypeAndGlobalNetwork(*s.GlobalNetworkId, resourceTypeLink, *s.LinkId),
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						In:  true,
						Out: true,
					},
				},
			},
		}

		items = append(items, &item)
	}

	return items, nil
}

//go:generate docgen ../../docs-data
// +overmind:type networkmanager-link
// +overmind:descriptiveType Networkmanager Link
// +overmind:get Get a Networkmanager Link
// +overmind:list List all Networkmanager Links
// +overmind:search Search for Networkmanager Links by GlobalNetworkId, or by GlobalNetworkId with SiteId
// +overmind:group AWS

func NewLinkSource(client *networkmanager.Client, accountID, region string) *sources.DescribeOnlySource[*networkmanager.GetLinksInput, *networkmanager.GetLinksOutput, *networkmanager.Client, *networkmanager.Options] {
	return &sources.DescribeOnlySource[*networkmanager.GetLinksInput, *networkmanager.GetLinksOutput, *networkmanager.Client, *networkmanager.Options]{
		Client:    client,
		AccountID: accountID,
		Region:    region,
		ItemType:  "networkmanager-link",
		DescribeFunc: func(ctx context.Context, client *networkmanager.Client, input *networkmanager.GetLinksInput) (*networkmanager.GetLinksOutput, error) {
			return client.GetLinks(ctx, input)
		},
		InputMapperGet: func(scope, query string) (*networkmanager.GetLinksInput, error) {
			// We are using a custom id of {globalNetworkId}|{linkId}
			sections := strings.Split(query, "|")

			if len(sections) != 2 {
				return nil, &sdp.QueryError{
					ErrorType:   sdp.QueryError_NOTFOUND,
					ErrorString: "invalid query for networkmanager-link get function",
				}
			}
			return &networkmanager.GetLinksInput{
				GlobalNetworkId: &sections[0],
				LinkIds: []string{
					sections[1],
				},
			}, nil
		},
		InputMapperList: func(scope string) (*networkmanager.GetLinksInput, error) {
			return nil, &sdp.QueryError{
				ErrorType:   sdp.QueryError_NOTFOUND,
				ErrorString: "list not supported for networkmanager-link, use search",
			}
		},
		PaginatorBuilder: func(client *networkmanager.Client, params *networkmanager.GetLinksInput) sources.Paginator[*networkmanager.GetLinksOutput, *networkmanager.Options] {
			return networkmanager.NewGetLinksPaginator(client, params)
		},
		OutputMapper: linkOutputMapper,
		InputMapperSearch: func(ctx context.Context, client *networkmanager.Client, scope, query string) (*networkmanager.GetLinksInput, error) {
			// We may search by only globalNetworkId or by using a custom id of {globalNetworkId}|{siteId}
			sections := strings.Split(query, "|")
			switch len(sections) {
			case 1:
				// globalNetworkId
				return &networkmanager.GetLinksInput{
					GlobalNetworkId: &sections[0],
				}, nil
			case 2:
				// {globalNetworkId}|{siteId}
				return &networkmanager.GetLinksInput{
					GlobalNetworkId: &sections[0],
					SiteId:          &sections[1],
				}, nil
			default:
				return nil, &sdp.QueryError{
					ErrorType:   sdp.QueryError_NOTFOUND,
					ErrorString: "invalid query for networkmanager-link get function",
				}
			}

		},
	}
}

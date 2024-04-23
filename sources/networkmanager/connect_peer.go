package networkmanager

import (
	"context"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func connectPeerGetFunc(ctx context.Context, client NetworkManagerClient, scope string, input *networkmanager.GetConnectPeerInput) (*sdp.Item, error) {
	out, err := client.GetConnectPeer(ctx, input)
	if err != nil {
		return nil, err
	}

	cn := out.ConnectPeer

	attributes, err := sources.ToAttributesCase(cn, "tags")

	if err != nil {
		return nil, err
	}

	item := sdp.Item{
		Type:            "networkmanager-connect-peer",
		UniqueAttribute: "connectPeerId",
		Attributes:      attributes,
		Scope:           scope,
		Tags:            tagsToMap(cn.Tags),
	}

	if cn.Configuration != nil {
		if cn.Configuration.CoreNetworkAddress != nil {
			//+overmind:link ip
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "ip",
					Method: sdp.QueryMethod_GET,
					Query:  *cn.Configuration.CoreNetworkAddress,
					Scope:  "global",
				},
				BlastPropagation: &sdp.BlastPropagation{
					In:  true,
					Out: true,
				},
			})
		}

		if cn.Configuration.PeerAddress != nil {
			//+overmind:link ip
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "ip",
					Method: sdp.QueryMethod_GET,
					Query:  *cn.Configuration.PeerAddress,
					Scope:  "global",
				},
				BlastPropagation: &sdp.BlastPropagation{
					In:  true,
					Out: true,
				},
			})
		}

		for _, config := range cn.Configuration.BgpConfigurations {
			if config.CoreNetworkAddress != nil {
				//+overmind:link ip
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "ip",
						Method: sdp.QueryMethod_GET,
						Query:  *config.CoreNetworkAddress,
						Scope:  "global",
					},
					BlastPropagation: &sdp.BlastPropagation{
						In:  true,
						Out: true,
					},
				})

				if config.PeerAddress != nil {
					//+overmind:link ip
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
						Query: &sdp.Query{
							Type:   "ip",
							Method: sdp.QueryMethod_GET,
							Query:  *config.PeerAddress,
							Scope:  "global",
						},
						BlastPropagation: &sdp.BlastPropagation{
							In:  true,
							Out: true,
						},
					})
				}

				if config.CoreNetworkAsn != nil {
					//+overmind:link rdap-asn
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
						Query: &sdp.Query{
							Type:   "rdap-asn",
							Method: sdp.QueryMethod_GET,
							Query:  strconv.FormatInt(*config.CoreNetworkAsn, 10),
							Scope:  "global",
						},
						BlastPropagation: &sdp.BlastPropagation{
							In:  true,
							Out: false,
						},
					})
				}

				if config.PeerAsn != nil {
					//+overmind:link rdap-asn
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
						Query: &sdp.Query{
							Type:   "rdap-asn",
							Method: sdp.QueryMethod_GET,
							Query:  strconv.FormatInt(*config.PeerAsn, 10),
							Scope:  "global",
						},
						BlastPropagation: &sdp.BlastPropagation{
							In:  true,
							Out: false,
						},
					})
				}
			}
		}
	}

	if cn.CoreNetworkId != nil {
		item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
			Query: &sdp.Query{
				// +overmind:link networkmanager-core-network
				Type:   "networkmanager-core-network",
				Method: sdp.QueryMethod_GET,
				Query:  *cn.CoreNetworkId,
				Scope:  scope,
			},
			BlastPropagation: &sdp.BlastPropagation{
				In:  true,
				Out: false,
			},
		})
	}

	if cn.SubnetArn != nil {
		if arn, err := sources.ParseARN(*cn.SubnetArn); err == nil {
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				//+overmind:link ec2-subnet
				Query: &sdp.Query{
					Type:   "ec2-subnet",
					Method: sdp.QueryMethod_SEARCH,
					Query:  *cn.SubnetArn,
					Scope:  sources.FormatScope(arn.AccountID, arn.Region),
				},
				BlastPropagation: &sdp.BlastPropagation{
					In:  true,
					Out: false,
				},
			})
		}
	}

	if cn.ConnectAttachmentId != nil {
		item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
			//+overmind:link networkmanager-connect-attachment
			Query: &sdp.Query{
				Type:   "networkmanager-connect-attachment",
				Method: sdp.QueryMethod_GET,
				Query:  *cn.ConnectAttachmentId,
				Scope:  scope,
			},
			BlastPropagation: &sdp.BlastPropagation{
				In:  true,
				Out: true,
			},
		})
	}

	switch cn.State {
	case types.ConnectPeerStateCreating:
		item.Health = sdp.Health_HEALTH_PENDING.Enum()
	case types.ConnectPeerStateFailed:
		item.Health = sdp.Health_HEALTH_ERROR.Enum()
	case types.ConnectPeerStateAvailable:
		item.Health = sdp.Health_HEALTH_OK.Enum()
	case types.ConnectPeerStateDeleting:
		item.Health = sdp.Health_HEALTH_PENDING.Enum()
	}

	return &item, nil
}

//go:generate docgen ../../docs-data
// +overmind:type networkmanager-connect-peer
// +overmind:descriptiveType Networkmanager Connect Peer
// +overmind:get Get a Networkmanager Connect Peer by id
// +overmind:group AWS
// +overmind:terraform:queryMap aws_networkmanager_connect_peer.id

func NewConnectPeerSource(client NetworkManagerClient, accountID, region string) *sources.AlwaysGetSource[*networkmanager.ListConnectPeersInput, *networkmanager.ListConnectPeersOutput, *networkmanager.GetConnectPeerInput, *networkmanager.GetConnectPeerOutput, NetworkManagerClient, *networkmanager.Options] {
	return &sources.AlwaysGetSource[*networkmanager.ListConnectPeersInput, *networkmanager.ListConnectPeersOutput, *networkmanager.GetConnectPeerInput, *networkmanager.GetConnectPeerOutput, NetworkManagerClient, *networkmanager.Options]{
		Client:    client,
		AccountID: accountID,
		Region:    region,
		ItemType:  "networkmanager-connect-peer",
		ListInput: &networkmanager.ListConnectPeersInput{},
		SearchInputMapper: func(scope, query string) (*networkmanager.ListConnectPeersInput, error) {
			// Search by CoreNetworkId
			return &networkmanager.ListConnectPeersInput{
				CoreNetworkId: &query,
			}, nil
		},
		GetInputMapper: func(scope, query string) *networkmanager.GetConnectPeerInput {
			return &networkmanager.GetConnectPeerInput{
				ConnectPeerId: &query,
			}
		},
		ListFuncPaginatorBuilder: func(client NetworkManagerClient, input *networkmanager.ListConnectPeersInput) sources.Paginator[*networkmanager.ListConnectPeersOutput, *networkmanager.Options] {
			return networkmanager.NewListConnectPeersPaginator(client, input)
		},
		ListFuncOutputMapper: func(output *networkmanager.ListConnectPeersOutput, input *networkmanager.ListConnectPeersInput) ([]*networkmanager.GetConnectPeerInput, error) {
			var inputs []*networkmanager.GetConnectPeerInput

			for _, connectPeer := range output.ConnectPeers {
				inputs = append(inputs, &networkmanager.GetConnectPeerInput{
					ConnectPeerId: connectPeer.ConnectPeerId,
				})
			}

			return inputs, nil

		},
		GetFunc: connectPeerGetFunc,
	}
}

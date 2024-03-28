package networkmanager

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func connectPeerGetFunc(ctx context.Context, client *networkmanager.Client, _, query string) (*types.ConnectPeer, error) {
	out, err := client.GetConnectPeer(ctx, &networkmanager.GetConnectPeerInput{
		ConnectPeerId: &query,
	})
	if err != nil {
		return nil, err
	}

	return out.ConnectPeer, nil
}

func connectPeerItemMapper(scope string, cn *types.ConnectPeer) (*sdp.Item, error) {
	attributes, err := sources.ToAttributesCase(cn)

	if err != nil {
		return nil, err
	}

	item := sdp.Item{
		Type:            "networkmanager-connect-peer",
		UniqueAttribute: "connectPeerId",
		Attributes:      attributes,
		Scope:           scope,
		LinkedItemQueries: []*sdp.LinkedItemQuery{
			{
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
			},
		},
	}

	return &item, nil
}

//go:generate docgen ../../docs-data
// +overmind:type networkmanager-connect-peer
// +overmind:descriptiveType Networkmanager Connect Peer
// +overmind:get Get a Networkmanager Connect Peer by id
// +overmind:group AWS
// +overmind:terraform:queryMap aws_networkmanager_connect_peer.core_network_id

func NewConnectPeerSource(client *networkmanager.Client, accountID, region string) *sources.GetListSource[*types.ConnectPeer, *networkmanager.Client, *networkmanager.Options] {
	return &sources.GetListSource[*types.ConnectPeer, *networkmanager.Client, *networkmanager.Options]{
		Client:    client,
		AccountID: accountID,
		Region:    region,
		ItemType:  "networkmanager-connect-peer",
		GetFunc: func(ctx context.Context, client *networkmanager.Client, scope string, query string) (*types.ConnectPeer, error) {
			return connectPeerGetFunc(ctx, client, scope, query)
		},
		ItemMapper: connectPeerItemMapper,

		ListFunc: func(ctx context.Context, client *networkmanager.Client, scope string) ([]*types.ConnectPeer, error) {
			out, err := client.ListConnectPeers(ctx, &networkmanager.ListConnectPeersInput{})
			if err != nil {
				return nil, err
			}

			connectPeers := make([]*types.ConnectPeer, len(out.ConnectPeers))

			for i, _ := range out.ConnectPeers {
				connectPeers[i] = &types.ConnectPeer{
					ConnectAttachmentId: out.ConnectPeers[i].ConnectAttachmentId,
					ConnectPeerId:       out.ConnectPeers[i].ConnectPeerId,
					CoreNetworkId:       out.ConnectPeers[i].CoreNetworkId,
					CreatedAt:           out.ConnectPeers[i].CreatedAt,
					EdgeLocation:        out.ConnectPeers[i].EdgeLocation,
					State:               out.ConnectPeers[i].ConnectPeerState,
					SubnetArn:           out.ConnectPeers[i].SubnetArn,
					Tags:                out.ConnectPeers[i].Tags,
				}
			}

			return connectPeers, nil
		},
		SearchFunc: func(ctx context.Context, client *networkmanager.Client, scope string, query string) ([]*types.ConnectPeer, error) {
			// Search by CoreNetworkId
			out, err := client.ListConnectPeers(ctx, &networkmanager.ListConnectPeersInput{
				CoreNetworkId: &query,
			})
			if err != nil {
				return nil, err
			}

			connectPeers := make([]*types.ConnectPeer, len(out.ConnectPeers))

			for i, _ := range out.ConnectPeers {
				connectPeers[i] = &types.ConnectPeer{
					ConnectAttachmentId: out.ConnectPeers[i].ConnectAttachmentId,
					ConnectPeerId:       out.ConnectPeers[i].ConnectPeerId,
					CoreNetworkId:       out.ConnectPeers[i].CoreNetworkId,
					CreatedAt:           out.ConnectPeers[i].CreatedAt,
					EdgeLocation:        out.ConnectPeers[i].EdgeLocation,
					State:               out.ConnectPeers[i].ConnectPeerState,
					SubnetArn:           out.ConnectPeers[i].SubnetArn,
					Tags:                out.ConnectPeers[i].Tags,
				}
			}

			return connectPeers, nil
		},
	}
}

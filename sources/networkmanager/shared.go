package networkmanager

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
)

type NetworkManagerClient interface {
	networkmanager.ListConnectPeersAPIClient
	networkmanager.ListCoreNetworksAPIClient

	GetConnectPeer(ctx context.Context, params *networkmanager.GetConnectPeerInput, optFns ...func(*networkmanager.Options)) (*networkmanager.GetConnectPeerOutput, error)
	GetCoreNetwork(ctx context.Context, params *networkmanager.GetCoreNetworkInput, optFns ...func(*networkmanager.Options)) (*networkmanager.GetCoreNetworkOutput, error)
}

// convertTags converts slice of ecs tags to a map
func tagsToMap(tags []types.Tag) map[string]string {
	tagsMap := make(map[string]string)

	for _, tag := range tags {
		if tag.Key != nil && tag.Value != nil {
			tagsMap[*tag.Key] = *tag.Value
		}
	}

	return tagsMap
}

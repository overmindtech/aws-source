package networkmanager

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
)

type NetworkmanagerClient interface {
	DescribeGlobalNetworks(ctx context.Context, params *networkmanager.DescribeGlobalNetworksInput, optFns ...func(*networkmanager.Options)) (*networkmanager.DescribeGlobalNetworksOutput, error)
	GetSites(ctx context.Context, params *networkmanager.GetSitesInput, optFns ...func(*networkmanager.Options)) (*networkmanager.GetSitesOutput, error)
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

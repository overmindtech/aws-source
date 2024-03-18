package networkmanager

import (
	"github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
)

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

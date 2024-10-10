package kms

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"
)

func tags(ctx context.Context, cli kmsClient, keyID string) (map[string]string, error) {
	if cli == nil {
		return nil, nil
	}

	output, err := cli.ListResourceTags(ctx, &kms.ListResourceTagsInput{
		KeyId: &keyID,
	})
	if err != nil {
		return nil, err
	}

	return tagsToMap(output.Tags), nil
}

func tagsToMap(tags []types.Tag) map[string]string {
	tagsMap := make(map[string]string)

	for _, tag := range tags {
		if tag.TagKey != nil && tag.TagValue != nil {
			tagsMap[*tag.TagKey] = *tag.TagValue
		}
	}

	return tagsMap
}

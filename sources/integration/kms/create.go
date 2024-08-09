package kms

import (
	"context"
	"errors"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/overmindtech/aws-source/sources/integration"
)

func createKMSKey(ctx context.Context, logger *slog.Logger, client *kms.Client, testID string) error {
	// check if a resource with the same tags already exists
	id, err := findActiveKeyIDByTags(ctx, client)
	if err != nil {
		if errors.As(err, new(integration.NotFoundError)) {
			logger.InfoContext(ctx, "Creating KMS key")
		} else {
			return err
		}
	}

	if id != nil {
		logger.InfoContext(ctx, "KMS key already exists")
		return nil
	}

	_, err = client.CreateKey(ctx, &kms.CreateKeyInput{
		Tags: resourceTags(keySrc, testID),
	})
	if err != nil {
		return err
	}

	return nil
}

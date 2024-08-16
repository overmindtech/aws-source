package kms

import (
	"context"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/overmindtech/aws-source/sources/integration"
)

const (
	keySrc   = "key"
	aliasSrc = "alias"
)

func setup(ctx context.Context, logger *slog.Logger, client *kms.Client) error {
	testID := integration.TestID()

	// Create KMS key
	keyID, err := createKey(ctx, logger, client, testID)
	if err != nil {
		return err
	}

	// Create KMS alias
	return createAlias(ctx, logger, client, *keyID)
}

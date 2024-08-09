package kms

import (
	"context"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/overmindtech/aws-source/sources/integration"
)

const keySrc = "key"

func setup(ctx context.Context, logger *slog.Logger, client *kms.Client) error {
	// Create KMS key
	return createKMSKey(ctx, logger, client, integration.TestID())
}

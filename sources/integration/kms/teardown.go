package kms

import (
	"context"
	"errors"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/overmindtech/aws-source/sources/integration"
)

func teardown(ctx context.Context, logger *slog.Logger, client *kms.Client) error {
	keyID, err := findActiveKeyIDByTags(ctx, client)
	if err != nil {
		nf := integration.NewNotFoundError(keySrc)
		if errors.As(err, &nf) {
			logger.WarnContext(ctx, "Key not found")
			return nil
		} else {
			return err
		}
	}

	return deleteKey(ctx, client, *keyID)
}

package kms

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/overmindtech/aws-source/sources/integration"
)

func createKey(ctx context.Context, logger *slog.Logger, client *kms.Client, testID string) (*string, error) {
	// check if a resource with the same tags already exists
	id, err := findActiveKeyIDByTags(ctx, client)
	if err != nil {
		if errors.As(err, new(integration.NotFoundError)) {
			logger.InfoContext(ctx, "Creating KMS key")
		} else {
			return nil, err
		}
	}

	if id != nil {
		logger.InfoContext(ctx, "KMS key already exists")
		return id, nil
	}

	response, err := client.CreateKey(ctx, &kms.CreateKeyInput{
		Tags: resourceTags(keySrc, testID),
	})
	if err != nil {
		return nil, err
	}

	return response.KeyMetadata.KeyId, nil
}

func createAlias(ctx context.Context, logger *slog.Logger, client *kms.Client, keyID string) error {
	aliasName := genAliasName()
	aliasNames, err := findAliasesByTargetKey(ctx, client, keyID)
	if err != nil {
		if nf := integration.NewNotFoundError(aliasSrc); errors.As(err, &nf) {
			logger.WarnContext(ctx, "Creating alias for the key", "keyID", keyID)
		} else {
			return err
		}
	}

	for _, aName := range aliasNames {
		if aName == aliasName {
			logger.InfoContext(ctx, "KMS alias already exists", "alias", aliasName, "keyID", keyID)
			return nil
		}
	}

	_, err = client.CreateAlias(ctx, &kms.CreateAliasInput{
		AliasName:   &aliasName,
		TargetKeyId: &keyID,
	})
	if err != nil {
		return err
	}

	return nil
}

func genAliasName() string {
	return fmt.Sprintf("alias/%s", integration.TestID())
}

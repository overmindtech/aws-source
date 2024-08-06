package kms

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/kms"
)

func deleteKey(ctx context.Context, client *kms.Client, keyID string) error {
	seven := int32(7)
	_, err := client.ScheduleKeyDeletion(ctx, &kms.ScheduleKeyDeletionInput{
		KeyId:               &keyID,
		PendingWindowInDays: &seven, // it can be minimum 7 days
	})
	return err
}

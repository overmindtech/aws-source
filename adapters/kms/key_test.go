package kms

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/overmindtech/aws-source/adapterhelpers"

	"github.com/aws/aws-sdk-go-v2/service/kms"
)

type testClient struct{}

func (t testClient) DescribeKey(ctx context.Context, params *kms.DescribeKeyInput, optFns ...func(*kms.Options)) (*kms.DescribeKeyOutput, error) {
	return &kms.DescribeKeyOutput{
		KeyMetadata: &types.KeyMetadata{
			AWSAccountId:          adapterhelpers.PtrString("846764612917"),
			KeyId:                 adapterhelpers.PtrString("b8a9477d-836c-491f-857e-07937918959b"),
			Arn:                   adapterhelpers.PtrString("arn:aws:kms:us-west-2:846764612917:key/b8a9477d-836c-491f-857e-07937918959b"),
			CreationDate:          adapterhelpers.PtrTime(time.Date(2017, 6, 30, 21, 44, 32, 140000000, time.UTC)),
			Enabled:               true,
			Description:           adapterhelpers.PtrString("Default KMS key that protects my S3 objects when no other key is defined"),
			KeyUsage:              types.KeyUsageTypeEncryptDecrypt,
			KeyState:              types.KeyStateEnabled,
			Origin:                types.OriginTypeAwsKms,
			KeyManager:            types.KeyManagerTypeAws,
			CustomerMasterKeySpec: types.CustomerMasterKeySpecSymmetricDefault,
			EncryptionAlgorithms: []types.EncryptionAlgorithmSpec{
				types.EncryptionAlgorithmSpecSymmetricDefault,
			},
		},
	}, nil
}

func (t testClient) ListKeys(context.Context, *kms.ListKeysInput, ...func(*kms.Options)) (*kms.ListKeysOutput, error) {
	return &kms.ListKeysOutput{
		Keys: []types.KeyListEntry{
			{
				KeyArn: adapterhelpers.PtrString("arn:aws:kms:us-west-2:111122223333:key/1234abcd-12ab-34cd-56ef-1234567890ab"),
				KeyId:  adapterhelpers.PtrString("1234abcd-12ab-34cd-56ef-1234567890ab"),
			},
			{
				KeyArn: adapterhelpers.PtrString("arn:aws:kms:us-west-2:111122223333:key/0987dcba-09fe-87dc-65ba-ab0987654321"),
				KeyId:  adapterhelpers.PtrString("0987dcba-09fe-87dc-65ba-ab0987654321"),
			},
			{
				KeyArn: adapterhelpers.PtrString("arn:aws:kms:us-east-2:111122223333:key/1a2b3c4d-5e6f-1a2b-3c4d-5e6f1a2b3c4d"),
				KeyId:  adapterhelpers.PtrString("1a2b3c4d-5e6f-1a2b-3c4d-5e6f1a2b3c4d"),
			},
		},
	}, nil
}

func (t testClient) ListResourceTags(context.Context, *kms.ListResourceTagsInput, ...func(*kms.Options)) (*kms.ListResourceTagsOutput, error) {
	return &kms.ListResourceTagsOutput{
		Tags: []types.Tag{
			{
				TagKey:   adapterhelpers.PtrString("Dept"),
				TagValue: adapterhelpers.PtrString("IT"),
			},
			{
				TagKey:   adapterhelpers.PtrString("Purpose"),
				TagValue: adapterhelpers.PtrString("Test"),
			},
			{
				TagKey:   adapterhelpers.PtrString("Name"),
				TagValue: adapterhelpers.PtrString("Test"),
			},
		},
	}, nil
}

func TestGetFunc(t *testing.T) {
	ctx := context.Background()
	cli := testClient{}

	item, err := getFunc(ctx, cli, "scope", &kms.DescribeKeyInput{
		KeyId: adapterhelpers.PtrString("1234abcd-12ab-34cd-56ef-1234567890ab"),
	})
	if err != nil {
		t.Fatal(err)
	}

	if err = item.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestNewKeyAdapter(t *testing.T) {
	config, account, region := adapterhelpers.GetAutoConfig(t)
	client := kms.NewFromConfig(config)

	adapter := NewKeyAdapter(client, account, region)

	test := adapterhelpers.E2ETest{
		Adapter: adapter,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}

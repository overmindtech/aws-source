package kms

import (
	"context"
	"testing"
	"time"

	"github.com/overmindtech/aws-source/adapters"
	"github.com/overmindtech/sdp-go"

	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"
)

func TestAliasOutputMapper(t *testing.T) {
	output := &kms.ListAliasesOutput{
		Aliases: []types.AliasListEntry{
			{
				AliasName:       adapters.PtrString("alias/test-key"),
				TargetKeyId:     adapters.PtrString("cf68415c-f4ae-48f2-87a7-3b52ce"),
				AliasArn:        adapters.PtrString("arn:aws:kms:us-west-2:123456789012:alias/test-key"),
				CreationDate:    adapters.PtrTime(time.Now()),
				LastUpdatedDate: adapters.PtrTime(time.Now()),
			},
		},
	}

	items, err := aliasOutputMapper(context.Background(), nil, "foo", nil, output)
	if err != nil {
		t.Fatal(err)
	}

	for _, item := range items {
		if err := item.Validate(); err != nil {
			t.Error(err)
		}
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %v", len(items))
	}

	item := items[0]

	tests := adapters.QueryTests{
		{
			ExpectedType:   "kms-key",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "cf68415c-f4ae-48f2-87a7-3b52ce",
			ExpectedScope:  "foo",
		},
	}

	tests.Execute(t, item)
}

func TestNewAliasAdapter(t *testing.T) {
	config, account, region := adapters.GetAutoConfig(t)
	client := kms.NewFromConfig(config)

	adapter := NewAliasAdapter(client, account, region)

	test := adapters.E2ETest{
		Adapter: adapter,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}

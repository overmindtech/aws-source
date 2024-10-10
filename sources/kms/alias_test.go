package kms

import (
	"context"
	"testing"
	"time"

	"github.com/overmindtech/sdp-go"

	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/overmindtech/aws-source/sources"
)

func TestAliasOutputMapper(t *testing.T) {
	output := &kms.ListAliasesOutput{
		Aliases: []types.AliasListEntry{
			{
				AliasName:       sources.PtrString("alias/test-key"),
				TargetKeyId:     sources.PtrString("cf68415c-f4ae-48f2-87a7-3b52ce"),
				AliasArn:        sources.PtrString("arn:aws:kms:us-west-2:123456789012:alias/test-key"),
				CreationDate:    sources.PtrTime(time.Now()),
				LastUpdatedDate: sources.PtrTime(time.Now()),
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

	tests := sources.QueryTests{
		{
			ExpectedType:   "kms-key",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "cf68415c-f4ae-48f2-87a7-3b52ce",
			ExpectedScope:  "foo",
		},
	}

	tests.Execute(t, item)
}

func TestNewAliasSource(t *testing.T) {
	config, account, region := sources.GetAutoConfig(t)
	client := kms.NewFromConfig(config)

	source := NewAliasSource(client, account, region)

	test := sources.E2ETest{
		Adapter: source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}

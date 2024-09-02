package kms

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

/*
An example list grants response:

{
    "Grants": [
        {
            "Constraints": {
                "EncryptionContextSubset": {
                    "aws:dynamodb:subscriberId": "123456789012",
                    "aws:dynamodb:tableName": "Services"
                }
            },
            "IssuingAccount": "arn:aws:iam::123456789012:root",
            "Name": "8276b9a6-6cf0-46f1-b2f0-7993a7f8c89a",
            "Operations": [
                "Decrypt",
                "Encrypt",
                "GenerateDataKey",
                "ReEncryptFrom",
                "ReEncryptTo",
                "RetireGrant",
                "DescribeKey"
            ],
            "GrantId": "1667b97d27cf748cf05b487217dd4179526c949d14fb3903858e25193253fe59",
            "KeyId": "arn:aws:kms:us-west-2:123456789012:key/1234abcd-12ab-34cd-56ef-1234567890ab",
            "RetiringPrincipal": "dynamodb.us-west-2.amazonaws.com",
            "GranteePrincipal": "dynamodb.us-west-2.amazonaws.com",
            "CreationDate": "2021-05-13T18:32:45.144000+00:00"
        }
    ]
}
*/

func TestGrantOutputMapper(t *testing.T) {
	output := &kms.ListGrantsOutput{
		Grants: []types.GrantListEntry{
			{
				Constraints: &types.GrantConstraints{
					EncryptionContextSubset: map[string]string{
						"aws:dynamodb:subscriberId": "123456789012",
						"aws:dynamodb:tableName":    "Services",
					},
				},
				IssuingAccount:    sources.PtrString("arn:aws:iam::123456789012:root"),
				Name:              sources.PtrString("8276b9a6-6cf0-46f1-b2f0-7993a7f8c89a"),
				Operations:        []types.GrantOperation{"Decrypt", "Encrypt", "GenerateDataKey", "ReEncryptFrom", "ReEncryptTo", "RetireGrant", "DescribeKey"},
				GrantId:           sources.PtrString("1667b97d27cf748cf05b487217dd4179526c949d14fb3903858e25193253fe59"),
				KeyId:             sources.PtrString("arn:aws:kms:us-west-2:123456789012:key/1234abcd-12ab-34cd-56ef-1234567890ab"),
				RetiringPrincipal: sources.PtrString("arn:aws:iam::account:role/role-name-with-path"),
				GranteePrincipal:  sources.PtrString("arn:aws:iam::account:user/user-name-with-path"),
				CreationDate:      sources.PtrTime(time.Now()),
			},
		},
	}

	items, err := grantOutputMapper(context.Background(), nil, "foo", nil, output)
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

	scope := sources.FormatScope("123456789012", "us-west-2")

	tests := sources.QueryTests{
		{
			ExpectedType:   "kms-key",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "1234abcd-12ab-34cd-56ef-1234567890ab",
			ExpectedScope:  scope,
		},
		{
			ExpectedType:   "iam-role",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "role-name-with-path",
			ExpectedScope:  scope,
		},
		{
			ExpectedType:   "iam-user",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "user-name-with-path",
			ExpectedScope:  scope,
		},
	}

	tests.Execute(t, item)
}

func TestNewGrantSource(t *testing.T) {
	config, account, region := sources.GetAutoConfig(t)
	client := kms.NewFromConfig(config)

	source := NewGrantSource(client, account, region)

	test := sources.E2ETest{
		Source:   source,
		Timeout:  10 * time.Second,
		SkipList: true,
	}

	test.Run(t)
}
package sns

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/overmindtech/aws-source/sources"
)

type mockDataProtectionPolicyClient struct{}

func (m mockDataProtectionPolicyClient) GetDataProtectionPolicy(ctx context.Context, params *sns.GetDataProtectionPolicyInput, optFns ...func(*sns.Options)) (*sns.GetDataProtectionPolicyOutput, error) {
	return &sns.GetDataProtectionPolicyOutput{
		DataProtectionPolicy: sources.PtrString("{\"Name\":\"data_protection_policy\",\"Description\":\"Example data protection policy\",\"Version\":\"2021-06-01\",\"Statement\":[{\"DataDirection\":\"Inbound\",\"Principal\":[\"*\"],\"DataIdentifier\":[\"arn:aws:dataprotection::aws:data-identifier/CreditCardNumber\"],\"Operation\":{\"Deny\":{}}}]}"),
	}, nil
}

func TestGetDataProtectionPolicyFunc(t *testing.T) {
	ctx := context.Background()
	cli := &mockDataProtectionPolicyClient{}

	item, err := getDataProtectionPolicyFunc(ctx, cli, "scope", &sns.GetDataProtectionPolicyInput{
		ResourceArn: sources.PtrString("arn:aws:sns:us-east-1:123456789012:mytopic"),
	})
	if err != nil {
		t.Fatal(err)
	}

	if err = item.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestNewDataProtectionPolicySource(t *testing.T) {
	config, account, region := sources.GetAutoConfig(t)

	source := NewDataProtectionPolicySource(config, account, region)

	test := sources.E2ETest{
		Source:   source,
		Timeout:  10 * time.Second,
		SkipList: true,
		SkipGet:  true,
	}

	test.Run(t)
}

package directconnect

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/directconnect"
	"github.com/aws/aws-sdk-go-v2/service/directconnect/types"
	"github.com/overmindtech/aws-source/adapters"
)

func TestCustomerMetadataOutputMapper(t *testing.T) {
	output := &directconnect.DescribeCustomerMetadataOutput{
		Agreements: []types.CustomerAgreement{
			{
				AgreementName: adapters.PtrString("example-customer-agreement"),
				Status:        adapters.PtrString("signed"),
			},
		},
	}

	items, err := customerMetadataOutputMapper(context.Background(), nil, "foo", nil, output)
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
}

func TestNewCustomerMetadataAdapter(t *testing.T) {
	client, account, region := GetAutoConfig(t)

	adapter := NewCustomerMetadataAdapter(client, account, region)

	test := adapters.E2ETest{
		Adapter: adapter,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}

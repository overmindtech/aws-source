package directconnect

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/directconnect"
	"github.com/aws/aws-sdk-go-v2/service/directconnect/types"
	"github.com/overmindtech/aws-source/sources"
)

func TestCustomerMetadataOutputMapper(t *testing.T) {
	output := &directconnect.DescribeCustomerMetadataOutput{
		Agreements: []types.CustomerAgreement{
			{
				AgreementName: sources.PtrString("example-customer-agreement"),
				Status:        sources.PtrString("signed"),
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

func TestNewCustomerMetadataSource(t *testing.T) {
	config, account, _ := sources.GetAutoConfig(t)

	source := NewCustomerMetadataSource(config, account, &TestRateLimit)

	test := sources.E2ETest{
		Source:  source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}

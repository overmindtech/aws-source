package directconnect

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/directconnect"
	"github.com/aws/aws-sdk-go-v2/service/directconnect/types"
	"github.com/overmindtech/aws-source/adapters"
	"github.com/overmindtech/sdp-go"
)

func TestHostedConnectionOutputMapper(t *testing.T) {
	output := &directconnect.DescribeHostedConnectionsOutput{
		Connections: []types.Connection{
			{
				AwsDeviceV2:          adapters.PtrString("EqDC2-123h49s71dabc"),
				AwsLogicalDeviceId:   adapters.PtrString("device-1"),
				Bandwidth:            adapters.PtrString("1Gbps"),
				ConnectionId:         adapters.PtrString("dxcon-fguhmqlc"),
				ConnectionName:       adapters.PtrString("My_Connection"),
				ConnectionState:      "down",
				EncryptionMode:       adapters.PtrString("must_encrypt"),
				HasLogicalRedundancy: "unknown",
				JumboFrameCapable:    adapters.PtrBool(true),
				LagId:                adapters.PtrString("dxlag-ffrz71kw"),
				LoaIssueTime:         adapters.PtrTime(time.Now()),
				Location:             adapters.PtrString("EqDC2"),
				Region:               adapters.PtrString("us-east-1"),
				ProviderName:         adapters.PtrString("provider-1"),
				OwnerAccount:         adapters.PtrString("123456789012"),
				PartnerName:          adapters.PtrString("partner-1"),
				Tags: []types.Tag{
					{
						Key:   adapters.PtrString("foo"),
						Value: adapters.PtrString("bar"),
					},
				},
			},
		},
	}

	items, err := hostedConnectionOutputMapper(context.Background(), nil, "foo", nil, output)
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
			ExpectedType:   "directconnect-lag",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "dxlag-ffrz71kw",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "directconnect-location",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "EqDC2",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "directconnect-loa",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "dxcon-fguhmqlc",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "directconnect-virtual-interface",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "dxcon-fguhmqlc",
			ExpectedScope:  "foo",
		},
	}

	tests.Execute(t, item)
}

func TestNewHostedConnectionSource(t *testing.T) {
	client, account, region := GetAutoConfig(t)

	source := NewHostedConnectionSource(client, account, region)

	test := adapters.E2ETest{
		Adapter:  source,
		Timeout:  10 * time.Second,
		SkipList: true,
	}

	test.Run(t)
}

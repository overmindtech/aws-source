package directconnect

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/directconnect"
	"github.com/aws/aws-sdk-go-v2/service/directconnect/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func TestConnectionOutputMapper(t *testing.T) {
	output := &directconnect.DescribeConnectionsOutput{
		Connections: []types.Connection{
			{
				AwsDeviceV2:          sources.PtrString("EqDC2-123h49s71dabc"),
				AwsLogicalDeviceId:   sources.PtrString("device-1"),
				Bandwidth:            sources.PtrString("1Gbps"),
				ConnectionId:         sources.PtrString("dxcon-fguhmqlc"),
				ConnectionName:       sources.PtrString("My_Connection"),
				ConnectionState:      "down",
				EncryptionMode:       sources.PtrString("must_encrypt"),
				HasLogicalRedundancy: "unknown",
				JumboFrameCapable:    sources.PtrBool(true),
				LagId:                sources.PtrString("dxlag-ffrz71kw"),
				LoaIssueTime:         sources.PtrTime(time.Now()),
				Location:             sources.PtrString("EqDC2"),
				Region:               sources.PtrString("us-east-1"),
				ProviderName:         sources.PtrString("provider-1"),
				OwnerAccount:         sources.PtrString("123456789012"),
				PartnerName:          sources.PtrString("partner-1"),
				Tags: []types.Tag{
					{
						Key:   sources.PtrString("foo"),
						Value: sources.PtrString("bar"),
					},
				},
			},
		},
	}

	items, err := connectionOutputMapper(context.Background(), nil, "foo", nil, output)
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

func TestNewConnectionSource(t *testing.T) {
	client, account, region := GetAutoConfig(t)

	source := NewConnectionSource(client, account, region)

	test := sources.E2ETest{
		Adapter: source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}

package networkmanager

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"github.com/overmindtech/aws-source/adapters"
	"github.com/overmindtech/sdp-go"
)

func (n NetworkManagerTestClient) GetConnectPeer(ctx context.Context, params *networkmanager.GetConnectPeerInput, optFns ...func(*networkmanager.Options)) (*networkmanager.GetConnectPeerOutput, error) {
	return &networkmanager.GetConnectPeerOutput{
		ConnectPeer: &types.ConnectPeer{
			Configuration: &types.ConnectPeerConfiguration{
				BgpConfigurations: []types.ConnectPeerBgpConfiguration{
					{
						CoreNetworkAddress: adapters.PtrString("1.4.2.4"),         // link
						CoreNetworkAsn:     adapters.PtrInt64(64512),              // link
						PeerAddress:        adapters.PtrString("123.123.123.123"), // link
						PeerAsn:            adapters.PtrInt64(64513),              // link
					},
				},
				CoreNetworkAddress: adapters.PtrString("1.1.1.3"),  // link
				PeerAddress:        adapters.PtrString("1.1.1.45"), // link
			},
			ConnectAttachmentId: adapters.PtrString("ca-1"), // link
			ConnectPeerId:       adapters.PtrString("cp-1"),
			CoreNetworkId:       adapters.PtrString("cn-1"), // link
			EdgeLocation:        adapters.PtrString("us-west-2"),
			State:               types.ConnectPeerStateAvailable,
			SubnetArn:           adapters.PtrString("arn:aws:ec2:us-west-2:123456789012:subnet/subnet-1"), // link
		},
	}, nil
}

func (n NetworkManagerTestClient) ListConnectPeers(context.Context, *networkmanager.ListConnectPeersInput, ...func(*networkmanager.Options)) (*networkmanager.ListConnectPeersOutput, error) {
	return nil, nil
}

func TestConnectPeerGetFunc(t *testing.T) {
	item, err := connectPeerGetFunc(context.Background(), NetworkManagerTestClient{}, "test", &networkmanager.GetConnectPeerInput{})
	if err != nil {
		t.Fatal(err)
	}

	// Ensure unique attribute
	err = item.Validate()
	if err != nil {
		t.Error(err)
	}

	if item.UniqueAttributeValue() != "cp-1" {
		t.Fatalf("expected cp-1, got %v", item.UniqueAttributeValue())
	}

	tests := adapters.QueryTests{
		{
			ExpectedType:   "ip",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "1.4.2.4",
			ExpectedScope:  "global",
		},
		{
			ExpectedType:   "ip",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "123.123.123.123",
			ExpectedScope:  "global",
		},
		{
			ExpectedType:   "rdap-asn",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "64512",
			ExpectedScope:  "global",
		},
		{
			ExpectedType:   "rdap-asn",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "64513",
			ExpectedScope:  "global",
		},
		{
			ExpectedType:   "ip",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "1.1.1.3",
			ExpectedScope:  "global",
		},
		{
			ExpectedType:   "ip",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "1.1.1.45",
			ExpectedScope:  "global",
		},
		{
			ExpectedType:   "networkmanager-core-network",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "cn-1",
			ExpectedScope:  "test",
		},
		{
			ExpectedType:   "networkmanager-connect-attachment",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "ca-1",
			ExpectedScope:  "test",
		},
		{
			ExpectedType:   "ec2-subnet",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "arn:aws:ec2:us-west-2:123456789012:subnet/subnet-1",
			ExpectedScope:  "123456789012.us-west-2",
		},
	}

	tests.Execute(t, item)
}

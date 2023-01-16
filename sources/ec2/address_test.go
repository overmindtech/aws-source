package ec2

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func TestAddressInputMapperGet(t *testing.T) {
	input, err := AddressInputMapperGet("foo", "az-name")

	if err != nil {
		t.Error(err)
	}

	if len(input.PublicIps) != 1 {
		t.Fatalf("expected 1 Address, got %v", len(input.PublicIps))
	}

	if input.PublicIps[0] != "az-name" {
		t.Errorf("expected Address to be to be az-name, got %v", input.PublicIps[0])
	}
}

func TestAddressInputMapperList(t *testing.T) {
	input, err := AddressInputMapperList("foo")

	if err != nil {
		t.Error(err)
	}

	if len(input.PublicIps) != 0 {
		t.Fatalf("expected 0 zone names, got %v", len(input.PublicIps))
	}
}

func TestAddressOutputMapper(t *testing.T) {
	output := ec2.DescribeAddressesOutput{
		Addresses: []types.Address{
			{
				PublicIp:           sources.PtrString("3.11.82.6"),
				AllocationId:       sources.PtrString("eipalloc-030a6f43bc6086267"),
				Domain:             types.DomainTypeVpc,
				PublicIpv4Pool:     sources.PtrString("amazon"),
				NetworkBorderGroup: sources.PtrString("eu-west-2"),
				InstanceId:         sources.PtrString("instance"),
				CarrierIp:          sources.PtrString("3.11.82.7"),
				CustomerOwnedIp:    sources.PtrString("3.11.82.8"),
				NetworkInterfaceId: sources.PtrString("foo"),
				PrivateIpAddress:   sources.PtrString("3.11.82.9"),
			},
		},
	}

	items, err := AddressOutputMapper("foo", &output)

	if err != nil {
		t.Error(err)
	}

	for _, item := range items {
		if err := item.Validate(); err != nil {
			t.Error(err)
		}
	}

	if len(items) != 1 {
		t.Errorf("expected 1 item, got %v", len(items))
	}

	item := items[0]

	// It doesn't really make sense to test anything other than the linked items
	// since the attributes are converted automatically
	tests := sources.ItemRequestTests{
		{
			ExpectedType:   "ip",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  *output.Addresses[0].PublicIp,
			ExpectedScope:  "global",
		},
		{
			ExpectedType:   "ec2-instance",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  *output.Addresses[0].InstanceId,
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "ip",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  *output.Addresses[0].CarrierIp,
			ExpectedScope:  "global",
		},
		{
			ExpectedType:   "ip",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  *output.Addresses[0].CustomerOwnedIp,
			ExpectedScope:  "global",
		},
		{
			ExpectedType:   "ec2-network-interface",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  *output.Addresses[0].NetworkInterfaceId,
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "ip",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  *output.Addresses[0].PrivateIpAddress,
			ExpectedScope:  "global",
		},
	}

	tests.Execute(t, item)
}

func TestNewAddressSource(t *testing.T) {
	config, account, _ := sources.GetAutoConfig(t)

	rateLimit := LimitBucket{
		MaxCapacity: 50,
		RefillRate:  10,
	}

	rateLimitCtx, rateLimitCancel := context.WithCancel(context.Background())
	defer rateLimitCancel()

	rateLimit.Start(rateLimitCtx)

	source := NewAddressSource(config, account, &rateLimit)

	test := sources.E2ETest{
		Source:  source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}

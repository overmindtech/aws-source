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

func TestVpcEndpointInputMapperGet(t *testing.T) {
	input, err := vpcEndpointInputMapperGet("foo", "bar")

	if err != nil {
		t.Error(err)
	}

	if len(input.VpcEndpointIds) != 1 {
		t.Fatalf("expected 1 VpcEndpoint ID, got %v", len(input.VpcEndpointIds))
	}

	if input.VpcEndpointIds[0] != "bar" {
		t.Errorf("expected VpcEndpoint ID to be bar, got %v", input.VpcEndpointIds[0])
	}
}

func TestVpcEndpointOutputMapper(t *testing.T) {
	output := &ec2.DescribeVpcEndpointsOutput{
		VpcEndpoints: []types.VpcEndpoint{
			{
				VpcEndpointId:     sources.PtrString("vpce-0d7892e00e573e701"),
				VpcEndpointType:   types.VpcEndpointTypeInterface,
				CreationTimestamp: sources.PtrTime(time.Now()),
				VpcId:             sources.PtrString("vpc-0d7892e00e573e701"), // link
				ServiceName:       sources.PtrString("com.amazonaws.us-east-1.s3"),
				State:             types.StateAvailable,
				PolicyDocument:    sources.PtrString("{\"Version\":\"2012-10-17\",\"Statement\":[{\"Action\":\"*\",\"Resource\":\"*\",\"Effect\":\"Allow\",\"Principal\":\"*\"},{\"Condition\":{\"StringNotEquals\":{\"aws:PrincipalAccount\":\"944651592624\"}},\"Action\":\"*\",\"Resource\":\"*\",\"Effect\":\"Deny\",\"Principal\":\"*\"}]}"), // parse this
				RouteTableIds: []string{
					"rtb-0d7892e00e573e701", // link
				},
				SubnetIds: []string{
					"subnet-0d7892e00e573e701", // link
				},
				Groups: []types.SecurityGroupIdentifier{
					{
						GroupId:   sources.PtrString("sg-0d7892e00e573e701"), // link
						GroupName: sources.PtrString("default"),
					},
				},
				IpAddressType:     types.IpAddressTypeIpv4,
				PrivateDnsEnabled: sources.PtrBool(true),
				RequesterManaged:  sources.PtrBool(false),
				DnsEntries: []types.DnsEntry{
					{
						DnsName:      sources.PtrString("vpce-0d7892e00e573e701-123456789012.us-east-1.vpce.amazonaws.com"), // link
						HostedZoneId: sources.PtrString("Z2F56UZL2M1ACD"),                                                   // link
					},
				},
				DnsOptions: &types.DnsOptions{
					DnsRecordIpType:                          types.DnsRecordIpTypeDualstack,
					PrivateDnsOnlyForInboundResolverEndpoint: sources.PtrBool(false),
				},
				LastError: &types.LastError{
					Code:    sources.PtrString("Client::ValidationException"),
					Message: sources.PtrString("The security group 'sg-0d7892e00e573e701' does not exist"),
				},
				NetworkInterfaceIds: []string{
					"eni-0d7892e00e573e701", // link
				},
				OwnerId: sources.PtrString("052392120703"),
				Tags: []types.Tag{
					{
						Key:   sources.PtrString("Name"),
						Value: sources.PtrString("my-vpce"),
					},
				},
			},
		},
	}

	items, err := vpcEndpointOutputMapper(context.Background(), nil, "foo", nil, output)

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

	tests := sources.QueryTests{
		{
			ExpectedType:   "ec2-vpc",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "vpc-0d7892e00e573e701",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "ec2-route-table",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "rtb-0d7892e00e573e701",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "ec2-subnet",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "subnet-0d7892e00e573e701",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "ec2-security-group",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "sg-0d7892e00e573e701",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "dns",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "vpce-0d7892e00e573e701-123456789012.us-east-1.vpce.amazonaws.com",
			ExpectedScope:  "global",
		},
		{
			ExpectedType:   "route53-hosted-zone",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "Z2F56UZL2M1ACD",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "ec2-network-interface",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "eni-0d7892e00e573e701",
			ExpectedScope:  "foo",
		},
	}

	tests.Execute(t, items[0])
}

func TestNewVpcEndpointSource(t *testing.T) {
	client, account, region := GetAutoConfig(t)

	source := NewVpcEndpointSource(client, account, region)

	test := sources.E2ETest{
		Source:  source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}

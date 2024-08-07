package elbv2

import (
	"context"
	"testing"
	"time"

	elbv2 "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func TestLoadBalancerOutputMapper(t *testing.T) {
	output := elbv2.DescribeLoadBalancersOutput{
		LoadBalancers: []types.LoadBalancer{
			{
				LoadBalancerArn:       sources.PtrString("arn:aws:elasticloadbalancing:eu-west-2:944651592624:loadbalancer/app/ingress/1bf10920c5bd199d"),
				DNSName:               sources.PtrString("ingress-1285969159.eu-west-2.elb.amazonaws.com"), // link
				CanonicalHostedZoneId: sources.PtrString("ZHURV8PSTC4K8"),                                  // link
				CreatedTime:           sources.PtrTime(time.Now()),
				LoadBalancerName:      sources.PtrString("ingress"),
				Scheme:                types.LoadBalancerSchemeEnumInternetFacing,
				VpcId:                 sources.PtrString("vpc-0c72199250cd479ea"), // link
				State: &types.LoadBalancerState{
					Code:   types.LoadBalancerStateEnumActive,
					Reason: sources.PtrString("reason"),
				},
				Type: types.LoadBalancerTypeEnumApplication,
				AvailabilityZones: []types.AvailabilityZone{
					{
						ZoneName: sources.PtrString("eu-west-2b"),               // link
						SubnetId: sources.PtrString("subnet-0960234bbc4edca03"), // link
						LoadBalancerAddresses: []types.LoadBalancerAddress{
							{
								AllocationId:       sources.PtrString("allocation-id"), // link?
								IPv6Address:        sources.PtrString(":::1"),          // link
								IpAddress:          sources.PtrString("1.1.1.1"),       // link
								PrivateIPv4Address: sources.PtrString("10.0.0.1"),      // link
							},
						},
						OutpostId: sources.PtrString("outpost-id"),
					},
				},
				SecurityGroups: []string{
					"sg-0b21edc8578ea3f93", // link
				},
				IpAddressType:         types.IpAddressTypeIpv4,
				CustomerOwnedIpv4Pool: sources.PtrString("ipv4-pool"), // link
			},
		},
	}

	items, err := loadBalancerOutputMapper(context.Background(), mockElbClient{}, "foo", nil, &output)

	if err != nil {
		t.Error(err)
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

	if item.GetTags()["foo"] != "bar" {
		t.Errorf("expected tag foo to be bar, got %v", item.GetTags()["foo"])
	}

	// It doesn't really make sense to test anything other than the linked items
	// since the attributes are converted automatically
	tests := sources.QueryTests{
		{
			ExpectedType:   "elbv2-target-group",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "arn:aws:elasticloadbalancing:eu-west-2:944651592624:loadbalancer/app/ingress/1bf10920c5bd199d",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "elbv2-listener",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "arn:aws:elasticloadbalancing:eu-west-2:944651592624:loadbalancer/app/ingress/1bf10920c5bd199d",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "dns",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "ingress-1285969159.eu-west-2.elb.amazonaws.com",
			ExpectedScope:  "global",
		},
		{
			ExpectedType:   "route53-hosted-zone",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "ZHURV8PSTC4K8",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "ec2-vpc",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "vpc-0c72199250cd479ea",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "ec2-subnet",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "subnet-0960234bbc4edca03",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "ec2-address",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "allocation-id",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "ip",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  ":::1",
			ExpectedScope:  "global",
		},
		{
			ExpectedType:   "ip",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "1.1.1.1",
			ExpectedScope:  "global",
		},
		{
			ExpectedType:   "ip",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "10.0.0.1",
			ExpectedScope:  "global",
		},
		{
			ExpectedType:   "ec2-security-group",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "sg-0b21edc8578ea3f93",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "ec2-coip-pool",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "ipv4-pool",
			ExpectedScope:  "foo",
		},
	}

	tests.Execute(t, item)
}

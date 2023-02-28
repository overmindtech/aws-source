package elb

import (
	"testing"
	"time"

	elb "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func TestLoadBalancerOutputMapper(t *testing.T) {
	output := &elb.DescribeLoadBalancersOutput{
		LoadBalancerDescriptions: []types.LoadBalancerDescription{
			{
				LoadBalancerName:          sources.PtrString("a8c3c8851f0df43fda89797c8e941a91"),
				DNSName:                   sources.PtrString("a8c3c8851f0df43fda89797c8e941a91-182843316.eu-west-2.elb.amazonaws.com"), // link
				CanonicalHostedZoneName:   sources.PtrString("a8c3c8851f0df43fda89797c8e941a91-182843316.eu-west-2.elb.amazonaws.com"), // link
				CanonicalHostedZoneNameID: sources.PtrString("ZHURV8PSTC4K8"),                                                          // link
				ListenerDescriptions: []types.ListenerDescription{
					{
						Listener: &types.Listener{
							Protocol:         sources.PtrString("TCP"),
							LoadBalancerPort: 7687,
							InstanceProtocol: sources.PtrString("TCP"),
							InstancePort:     30133,
						},
						PolicyNames: []string{},
					},
					{
						Listener: &types.Listener{
							Protocol:         sources.PtrString("TCP"),
							LoadBalancerPort: 7473,
							InstanceProtocol: sources.PtrString("TCP"),
							InstancePort:     31459,
						},
						PolicyNames: []string{},
					},
					{
						Listener: &types.Listener{
							Protocol:         sources.PtrString("TCP"),
							LoadBalancerPort: 7474,
							InstanceProtocol: sources.PtrString("TCP"),
							InstancePort:     30761,
						},
						PolicyNames: []string{},
					},
				},
				Policies: &types.Policies{
					AppCookieStickinessPolicies: []types.AppCookieStickinessPolicy{
						{
							CookieName: sources.PtrString("foo"),
							PolicyName: sources.PtrString("policy"),
						},
					},
					LBCookieStickinessPolicies: []types.LBCookieStickinessPolicy{
						{
							CookieExpirationPeriod: sources.PtrInt64(10),
							PolicyName:             sources.PtrString("name"),
						},
					},
					OtherPolicies: []string{},
				},
				BackendServerDescriptions: []types.BackendServerDescription{
					{
						InstancePort: 443,
						PolicyNames:  []string{},
					},
				},
				AvailabilityZones: []string{ // link
					"euwest-2b",
					"euwest-2a",
					"euwest-2c",
				},
				Subnets: []string{ // link
					"subnet0960234bbc4edca03",
					"subnet09d5f6fa75b0b4569",
					"subnet0e234bef35fc4a9e1",
				},
				VPCId: sources.PtrString("vpc-0c72199250cd479ea"), // link
				Instances: []types.Instance{
					{
						InstanceId: sources.PtrString("i-0337802d908b4a81e"), // link *2 to ec2-instance and health
					},
				},
				HealthCheck: &types.HealthCheck{
					Target:             sources.PtrString("HTTP:31151/healthz"),
					Interval:           10,
					Timeout:            5,
					UnhealthyThreshold: 6,
					HealthyThreshold:   2,
				},
				SourceSecurityGroup: &types.SourceSecurityGroup{
					OwnerAlias: sources.PtrString("944651592624"),
					GroupName:  sources.PtrString("k8s-elb-a8c3c8851f0df43fda89797c8e941a91"), // link
				},
				SecurityGroups: []string{
					"sg097e3cfdfc6d53b77", // link
				},
				CreatedTime: sources.PtrTime(time.Now()),
				Scheme:      sources.PtrString("internet-facing"),
			},
		},
	}

	items, err := LoadBalancerOutputMapper("foo", nil, output)

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

	// It doesn't really make sense to test anything other than the linked items
	// since the attributes are converted automatically
	tests := sources.ItemRequestTests{
		{
			ExpectedType:   "dns",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  "a8c3c8851f0df43fda89797c8e941a91-182843316.eu-west-2.elb.amazonaws.com",
			ExpectedScope:  "global",
		},
		{
			ExpectedType:   "dns",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  "a8c3c8851f0df43fda89797c8e941a91-182843316.eu-west-2.elb.amazonaws.com",
			ExpectedScope:  "global",
		},
		{
			ExpectedType:   "route53-hosted-zone",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  "ZHURV8PSTC4K8",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "ec2-availability-zone",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  "euwest-2b",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "ec2-availability-zone",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  "euwest-2a",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "ec2-availability-zone",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  "euwest-2c",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "ec2-subnet",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  "subnet0960234bbc4edca03",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "ec2-subnet",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  "subnet09d5f6fa75b0b4569",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "ec2-subnet",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  "subnet0e234bef35fc4a9e1",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "ec2-vpc",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  "vpc-0c72199250cd479ea",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "ec2-instance",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  "i-0337802d908b4a81e",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "elb-instance-health",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  "a8c3c8851f0df43fda89797c8e941a91/i-0337802d908b4a81e",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "ec2-security-group",
			ExpectedMethod: sdp.RequestMethod_SEARCH,
			ExpectedQuery:  "k8s-elb-a8c3c8851f0df43fda89797c8e941a91",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "ec2-security-group",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  "sg097e3cfdfc6d53b77",
			ExpectedScope:  "foo",
		},
	}

	tests.Execute(t, item)
}

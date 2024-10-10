package elbv2

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	elbv2 "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/overmindtech/aws-source/adapters"
	"github.com/overmindtech/sdp-go"
)

func TestTargetGroupOutputMapper(t *testing.T) {
	output := elbv2.DescribeTargetGroupsOutput{
		TargetGroups: []types.TargetGroup{
			{
				TargetGroupArn:             adapters.PtrString("arn:aws:elasticloadbalancing:eu-west-2:944651592624:targetgroup/k8s-default-apiserve-d87e8f7010/559d207158e41222"),
				TargetGroupName:            adapters.PtrString("k8s-default-apiserve-d87e8f7010"),
				Protocol:                   types.ProtocolEnumHttp,
				Port:                       adapters.PtrInt32(8080),
				VpcId:                      adapters.PtrString("vpc-0c72199250cd479ea"), // link
				HealthCheckProtocol:        types.ProtocolEnumHttp,
				HealthCheckPort:            adapters.PtrString("traffic-port"),
				HealthCheckEnabled:         adapters.PtrBool(true),
				HealthCheckIntervalSeconds: adapters.PtrInt32(10),
				HealthCheckTimeoutSeconds:  adapters.PtrInt32(10),
				HealthyThresholdCount:      adapters.PtrInt32(10),
				UnhealthyThresholdCount:    adapters.PtrInt32(10),
				HealthCheckPath:            adapters.PtrString("/"),
				Matcher: &types.Matcher{
					HttpCode: adapters.PtrString("200"),
					GrpcCode: adapters.PtrString("code"),
				},
				LoadBalancerArns: []string{
					"arn:aws:elasticloadbalancing:eu-west-2:944651592624:loadbalancer/app/ingress/1bf10920c5bd199d", // link
				},
				TargetType:      types.TargetTypeEnumIp,
				ProtocolVersion: adapters.PtrString("HTTP1"),
				IpAddressType:   types.TargetGroupIpAddressTypeEnumIpv4,
			},
		},
	}

	items, err := targetGroupOutputMapper(context.Background(), mockElbClient{}, "foo", nil, &output)

	if err != nil {
		t.Error(err)
	}

	for _, item := range items {
		if item.GetTags()["foo"] != "bar" {
			t.Errorf("expected tag foo to be bar, got %v", item.GetTags()["foo"])
		}

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
	tests := adapters.QueryTests{
		{
			ExpectedType:   "ec2-vpc",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "vpc-0c72199250cd479ea",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "elbv2-load-balancer",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "arn:aws:elasticloadbalancing:eu-west-2:944651592624:loadbalancer/app/ingress/1bf10920c5bd199d",
			ExpectedScope:  "944651592624.eu-west-2",
		},
	}

	tests.Execute(t, item)
}

func TestNewTargetGroupAdapter(t *testing.T) {
	config, account, region := adapters.GetAutoConfig(t)
	client := elasticloadbalancingv2.NewFromConfig(config)

	adapter := NewTargetGroupAdapter(client, account, region)

	test := adapters.E2ETest{
		Adapter: adapter,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}

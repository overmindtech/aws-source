package autoscaling

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling/types"
	"github.com/overmindtech/aws-source/adapters"
	"github.com/overmindtech/sdp-go"
)

func TestAutoScalingGroupOutputMapper(t *testing.T) {
	t.Parallel()

	output := autoscaling.DescribeAutoScalingGroupsOutput{
		AutoScalingGroups: []types.AutoScalingGroup{
			{
				AutoScalingGroupName: adapters.PtrString("eks-default-20230117110031319900000013-96c2dfb1-a11b-b5e4-6efb-0fea7e22855c"),
				AutoScalingGroupARN:  adapters.PtrString("arn:aws:autoscaling:eu-west-2:944651592624:autoScalingGroup:1cbb0e22-818f-4d8b-8662-77f73d3713ca:autoScalingGroupName/eks-default-20230117110031319900000013-96c2dfb1-a11b-b5e4-6efb-0fea7e22855c"),
				MixedInstancesPolicy: &types.MixedInstancesPolicy{
					LaunchTemplate: &types.LaunchTemplate{
						LaunchTemplateSpecification: &types.LaunchTemplateSpecification{
							LaunchTemplateId:   adapters.PtrString("lt-0174ff2b8909d0c75"), // link
							LaunchTemplateName: adapters.PtrString("eks-96c2dfb1-a11b-b5e4-6efb-0fea7e22855c"),
							Version:            adapters.PtrString("1"),
						},
						Overrides: []types.LaunchTemplateOverrides{
							{
								InstanceType: adapters.PtrString("t3.large"),
							},
						},
					},
					InstancesDistribution: &types.InstancesDistribution{
						OnDemandAllocationStrategy:          adapters.PtrString("prioritized"),
						OnDemandBaseCapacity:                adapters.PtrInt32(0),
						OnDemandPercentageAboveBaseCapacity: adapters.PtrInt32(100),
						SpotAllocationStrategy:              adapters.PtrString("lowest-price"),
						SpotInstancePools:                   adapters.PtrInt32(2),
					},
				},
				MinSize:         adapters.PtrInt32(1),
				MaxSize:         adapters.PtrInt32(3),
				DesiredCapacity: adapters.PtrInt32(1),
				DefaultCooldown: adapters.PtrInt32(300),
				AvailabilityZones: []string{ // link
					"eu-west-2c",
					"eu-west-2a",
					"eu-west-2b",
				},
				LoadBalancerNames: []string{}, // Ignored, classic load balancer
				TargetGroupARNs: []string{
					"arn:partition:service:region:account-id:resource-type/resource-id", // link
				},
				HealthCheckType:        adapters.PtrString("EC2"),
				HealthCheckGracePeriod: adapters.PtrInt32(15),
				Instances: []types.Instance{
					{
						InstanceId:       adapters.PtrString("i-0be6c4fe789cb1b78"), // link
						InstanceType:     adapters.PtrString("t3.large"),
						AvailabilityZone: adapters.PtrString("eu-west-2c"),
						LifecycleState:   types.LifecycleStateInService,
						HealthStatus:     adapters.PtrString("Healthy"),
						LaunchTemplate: &types.LaunchTemplateSpecification{
							LaunchTemplateId:   adapters.PtrString("lt-0174ff2b8909d0c75"), // Link
							LaunchTemplateName: adapters.PtrString("eks-96c2dfb1-a11b-b5e4-6efb-0fea7e22855c"),
							Version:            adapters.PtrString("1"),
						},
						ProtectedFromScaleIn: adapters.PtrBool(false),
					},
				},
				CreatedTime:        adapters.PtrTime(time.Now()),
				SuspendedProcesses: []types.SuspendedProcess{},
				VPCZoneIdentifier:  adapters.PtrString("subnet-0e234bef35fc4a9e1,subnet-09d5f6fa75b0b4569,subnet-0960234bbc4edca03"),
				EnabledMetrics:     []types.EnabledMetric{},
				Tags: []types.TagDescription{
					{
						ResourceId:        adapters.PtrString("eks-default-20230117110031319900000013-96c2dfb1-a11b-b5e4-6efb-0fea7e22855c"),
						ResourceType:      adapters.PtrString("auto-scaling-group"),
						Key:               adapters.PtrString("eks:cluster-name"),
						Value:             adapters.PtrString("dogfood"),
						PropagateAtLaunch: adapters.PtrBool(true),
					},
					{
						ResourceId:        adapters.PtrString("eks-default-20230117110031319900000013-96c2dfb1-a11b-b5e4-6efb-0fea7e22855c"),
						ResourceType:      adapters.PtrString("auto-scaling-group"),
						Key:               adapters.PtrString("eks:nodegroup-name"),
						Value:             adapters.PtrString("default-20230117110031319900000013"),
						PropagateAtLaunch: adapters.PtrBool(true),
					},
					{
						ResourceId:        adapters.PtrString("eks-default-20230117110031319900000013-96c2dfb1-a11b-b5e4-6efb-0fea7e22855c"),
						ResourceType:      adapters.PtrString("auto-scaling-group"),
						Key:               adapters.PtrString("k8s.io/cluster-autoscaler/dogfood"),
						Value:             adapters.PtrString("owned"),
						PropagateAtLaunch: adapters.PtrBool(true),
					},
					{
						ResourceId:        adapters.PtrString("eks-default-20230117110031319900000013-96c2dfb1-a11b-b5e4-6efb-0fea7e22855c"),
						ResourceType:      adapters.PtrString("auto-scaling-group"),
						Key:               adapters.PtrString("k8s.io/cluster-autoscaler/enabled"),
						Value:             adapters.PtrString("true"),
						PropagateAtLaunch: adapters.PtrBool(true),
					},
					{
						ResourceId:        adapters.PtrString("eks-default-20230117110031319900000013-96c2dfb1-a11b-b5e4-6efb-0fea7e22855c"),
						ResourceType:      adapters.PtrString("auto-scaling-group"),
						Key:               adapters.PtrString("kubernetes.io/cluster/dogfood"),
						Value:             adapters.PtrString("owned"),
						PropagateAtLaunch: adapters.PtrBool(true),
					},
				},
				TerminationPolicies: []string{
					"AllocationStrategy",
					"OldestLaunchTemplate",
					"OldestInstance",
				},
				NewInstancesProtectedFromScaleIn: adapters.PtrBool(false),
				ServiceLinkedRoleARN:             adapters.PtrString("arn:aws:iam::944651592624:role/aws-service-role/autoscaling.amazonaws.com/AWSServiceRoleForAutoScaling"), // link
				CapacityRebalance:                adapters.PtrBool(true),
				TrafficSources: []types.TrafficSourceIdentifier{
					{
						Identifier: adapters.PtrString("arn:partition:service:region:account-id:resource-type/resource-id"), // We will skip this for now since it's related to VPC lattice groups which are still in preview
					},
				},
				Context:                 adapters.PtrString("foo"),
				DefaultInstanceWarmup:   adapters.PtrInt32(10),
				DesiredCapacityType:     adapters.PtrString("foo"),
				LaunchConfigurationName: adapters.PtrString("launchConfig"), // link
				LaunchTemplate: &types.LaunchTemplateSpecification{
					LaunchTemplateId:   adapters.PtrString("id"), // link
					LaunchTemplateName: adapters.PtrString("launchTemplateName"),
				},
				MaxInstanceLifetime: adapters.PtrInt32(30),
				PlacementGroup:      adapters.PtrString("placementGroup"), // link (ec2)
				PredictedCapacity:   adapters.PtrInt32(1),
				Status:              adapters.PtrString("OK"),
				WarmPoolConfiguration: &types.WarmPoolConfiguration{
					InstanceReusePolicy: &types.InstanceReusePolicy{
						ReuseOnScaleIn: adapters.PtrBool(true),
					},
					MaxGroupPreparedCapacity: adapters.PtrInt32(1),
					MinSize:                  adapters.PtrInt32(1),
					PoolState:                types.WarmPoolStateHibernated,
					Status:                   types.WarmPoolStatusPendingDelete,
				},
				WarmPoolSize: adapters.PtrInt32(1),
			},
		},
	}

	items, err := autoScalingGroupOutputMapper(context.Background(), nil, "foo", nil, &output)

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
	tests := adapters.QueryTests{
		{
			ExpectedType:   "ec2-launch-template",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "lt-0174ff2b8909d0c75",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "elbv2-target-group",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "arn:partition:service:region:account-id:resource-type/resource-id",
			ExpectedScope:  "account-id.region",
		},
		{
			ExpectedType:   "ec2-instance",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "i-0be6c4fe789cb1b78",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "iam-role",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "arn:aws:iam::944651592624:role/aws-service-role/autoscaling.amazonaws.com/AWSServiceRoleForAutoScaling",
			ExpectedScope:  "944651592624",
		},
		{
			ExpectedType:   "autoscaling-launch-configuration",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "launchConfig",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "ec2-launch-template",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "id",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "ec2-placement-group",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "placementGroup",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "ec2-launch-template",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "lt-0174ff2b8909d0c75",
			ExpectedScope:  "foo",
		},
	}

	tests.Execute(t, item)
}

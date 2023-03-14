package autoscaling

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func TestAutoScalingGroupOutputMapper(t *testing.T) {
	t.Parallel()

	output := autoscaling.DescribeAutoScalingGroupsOutput{
		AutoScalingGroups: []types.AutoScalingGroup{
			{
				AutoScalingGroupName: sources.PtrString("eks-default-20230117110031319900000013-96c2dfb1-a11b-b5e4-6efb-0fea7e22855c"),
				AutoScalingGroupARN:  sources.PtrString("arn:aws:autoscaling:eu-west-2:944651592624:autoScalingGroup:1cbb0e22-818f-4d8b-8662-77f73d3713ca:autoScalingGroupName/eks-default-20230117110031319900000013-96c2dfb1-a11b-b5e4-6efb-0fea7e22855c"),
				MixedInstancesPolicy: &types.MixedInstancesPolicy{
					LaunchTemplate: &types.LaunchTemplate{
						LaunchTemplateSpecification: &types.LaunchTemplateSpecification{
							LaunchTemplateId:   sources.PtrString("lt-0174ff2b8909d0c75"), // link
							LaunchTemplateName: sources.PtrString("eks-96c2dfb1-a11b-b5e4-6efb-0fea7e22855c"),
							Version:            sources.PtrString("1"),
						},
						Overrides: []types.LaunchTemplateOverrides{
							{
								InstanceType: sources.PtrString("t3.large"),
							},
						},
					},
					InstancesDistribution: &types.InstancesDistribution{
						OnDemandAllocationStrategy:          sources.PtrString("prioritized"),
						OnDemandBaseCapacity:                sources.PtrInt32(0),
						OnDemandPercentageAboveBaseCapacity: sources.PtrInt32(100),
						SpotAllocationStrategy:              sources.PtrString("lowest-price"),
						SpotInstancePools:                   sources.PtrInt32(2),
					},
				},
				MinSize:         sources.PtrInt32(1),
				MaxSize:         sources.PtrInt32(3),
				DesiredCapacity: sources.PtrInt32(1),
				DefaultCooldown: sources.PtrInt32(300),
				AvailabilityZones: []string{
					"eu-west-2c",
					"eu-west-2a",
					"eu-west-2b",
				},
				LoadBalancerNames: []string{}, // Ignored, classic load balancer
				TargetGroupARNs: []string{
					"arn:partition:service:region:account-id:resource-type/resource-id", // link
				},
				HealthCheckType:        sources.PtrString("EC2"),
				HealthCheckGracePeriod: sources.PtrInt32(15),
				Instances: []types.Instance{
					{
						InstanceId:       sources.PtrString("i-0be6c4fe789cb1b78"), // link
						InstanceType:     sources.PtrString("t3.large"),
						AvailabilityZone: sources.PtrString("eu-west-2c"),
						LifecycleState:   types.LifecycleStateInService,
						HealthStatus:     sources.PtrString("Healthy"),
						LaunchTemplate: &types.LaunchTemplateSpecification{
							LaunchTemplateId:   sources.PtrString("lt-0174ff2b8909d0c75"), // Link
							LaunchTemplateName: sources.PtrString("eks-96c2dfb1-a11b-b5e4-6efb-0fea7e22855c"),
							Version:            sources.PtrString("1"),
						},
						ProtectedFromScaleIn: sources.PtrBool(false),
					},
				},
				CreatedTime:        sources.PtrTime(time.Now()),
				SuspendedProcesses: []types.SuspendedProcess{},
				VPCZoneIdentifier:  sources.PtrString("subnet-0e234bef35fc4a9e1,subnet-09d5f6fa75b0b4569,subnet-0960234bbc4edca03"),
				EnabledMetrics:     []types.EnabledMetric{},
				Tags: []types.TagDescription{
					{
						ResourceId:        sources.PtrString("eks-default-20230117110031319900000013-96c2dfb1-a11b-b5e4-6efb-0fea7e22855c"),
						ResourceType:      sources.PtrString("auto-scaling-group"),
						Key:               sources.PtrString("eks:cluster-name"),
						Value:             sources.PtrString("dogfood"),
						PropagateAtLaunch: sources.PtrBool(true),
					},
					{
						ResourceId:        sources.PtrString("eks-default-20230117110031319900000013-96c2dfb1-a11b-b5e4-6efb-0fea7e22855c"),
						ResourceType:      sources.PtrString("auto-scaling-group"),
						Key:               sources.PtrString("eks:nodegroup-name"),
						Value:             sources.PtrString("default-20230117110031319900000013"),
						PropagateAtLaunch: sources.PtrBool(true),
					},
					{
						ResourceId:        sources.PtrString("eks-default-20230117110031319900000013-96c2dfb1-a11b-b5e4-6efb-0fea7e22855c"),
						ResourceType:      sources.PtrString("auto-scaling-group"),
						Key:               sources.PtrString("k8s.io/cluster-autoscaler/dogfood"),
						Value:             sources.PtrString("owned"),
						PropagateAtLaunch: sources.PtrBool(true),
					},
					{
						ResourceId:        sources.PtrString("eks-default-20230117110031319900000013-96c2dfb1-a11b-b5e4-6efb-0fea7e22855c"),
						ResourceType:      sources.PtrString("auto-scaling-group"),
						Key:               sources.PtrString("k8s.io/cluster-autoscaler/enabled"),
						Value:             sources.PtrString("true"),
						PropagateAtLaunch: sources.PtrBool(true),
					},
					{
						ResourceId:        sources.PtrString("eks-default-20230117110031319900000013-96c2dfb1-a11b-b5e4-6efb-0fea7e22855c"),
						ResourceType:      sources.PtrString("auto-scaling-group"),
						Key:               sources.PtrString("kubernetes.io/cluster/dogfood"),
						Value:             sources.PtrString("owned"),
						PropagateAtLaunch: sources.PtrBool(true),
					},
				},
				TerminationPolicies: []string{
					"AllocationStrategy",
					"OldestLaunchTemplate",
					"OldestInstance",
				},
				NewInstancesProtectedFromScaleIn: sources.PtrBool(false),
				ServiceLinkedRoleARN:             sources.PtrString("arn:aws:iam::944651592624:role/aws-service-role/autoscaling.amazonaws.com/AWSServiceRoleForAutoScaling"), // link
				CapacityRebalance:                sources.PtrBool(true),
				TrafficSources: []types.TrafficSourceIdentifier{
					{
						Identifier: sources.PtrString("arn:partition:service:region:account-id:resource-type/resource-id"), // We will skip this for now since it's related to VPC lattice groups which are still in preview
					},
				},
				Context:                 sources.PtrString("foo"),
				DefaultInstanceWarmup:   sources.PtrInt32(10),
				DesiredCapacityType:     sources.PtrString("foo"),
				LaunchConfigurationName: sources.PtrString("launchConfig"), // link
				LaunchTemplate: &types.LaunchTemplateSpecification{
					LaunchTemplateId:   sources.PtrString("id"), // link
					LaunchTemplateName: sources.PtrString("launchTemplateName"),
				},
				MaxInstanceLifetime: sources.PtrInt32(30),
				PlacementGroup:      sources.PtrString("placementGroup"), // link (ec2)
				PredictedCapacity:   sources.PtrInt32(1),
				Status:              sources.PtrString("OK"),
				WarmPoolConfiguration: &types.WarmPoolConfiguration{
					InstanceReusePolicy: &types.InstanceReusePolicy{
						ReuseOnScaleIn: sources.PtrBool(true),
					},
					MaxGroupPreparedCapacity: sources.PtrInt32(1),
					MinSize:                  sources.PtrInt32(1),
					PoolState:                types.WarmPoolStateHibernated,
					Status:                   types.WarmPoolStatusPendingDelete,
				},
				WarmPoolSize: sources.PtrInt32(1),
			},
		},
	}

	items, err := autoScalingGroupOutputMapper("foo", nil, &output)

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

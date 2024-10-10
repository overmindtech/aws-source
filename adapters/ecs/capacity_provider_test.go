package ecs

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/overmindtech/aws-source/adapters"
	"github.com/overmindtech/sdp-go"
)

func (t *TestClient) DescribeCapacityProviders(ctx context.Context, params *ecs.DescribeCapacityProvidersInput, optFns ...func(*ecs.Options)) (*ecs.DescribeCapacityProvidersOutput, error) {
	pages := map[string]*ecs.DescribeCapacityProvidersOutput{
		"": {
			CapacityProviders: []types.CapacityProvider{
				{
					CapacityProviderArn: adapters.PtrString("arn:aws:ecs:eu-west-2:052392120703:capacity-provider/FARGATE"),
					Name:                adapters.PtrString("FARGATE"),
					Status:              types.CapacityProviderStatusActive,
				},
			},
			NextToken: adapters.PtrString("one"),
		},
		"one": {
			CapacityProviders: []types.CapacityProvider{
				{
					CapacityProviderArn: adapters.PtrString("arn:aws:ecs:eu-west-2:052392120703:capacity-provider/FARGATE_SPOT"),
					Name:                adapters.PtrString("FARGATE_SPOT"),
					Status:              types.CapacityProviderStatusActive,
				},
			},
			NextToken: adapters.PtrString("two"),
		},
		"two": {
			CapacityProviders: []types.CapacityProvider{
				{
					CapacityProviderArn: adapters.PtrString("arn:aws:ecs:eu-west-2:052392120703:capacity-provider/test"),
					Name:                adapters.PtrString("test"),
					Status:              types.CapacityProviderStatusActive,
					AutoScalingGroupProvider: &types.AutoScalingGroupProvider{
						AutoScalingGroupArn: adapters.PtrString("arn:aws:autoscaling:eu-west-2:052392120703:autoScalingGroup:9df90815-98c1-4136-a12a-90abef1c4e4e:autoScalingGroupName/ecs-test"),
						ManagedScaling: &types.ManagedScaling{
							Status:                 types.ManagedScalingStatusEnabled,
							TargetCapacity:         adapters.PtrInt32(80),
							MinimumScalingStepSize: adapters.PtrInt32(1),
							MaximumScalingStepSize: adapters.PtrInt32(10000),
							InstanceWarmupPeriod:   adapters.PtrInt32(300),
						},
						ManagedTerminationProtection: types.ManagedTerminationProtectionDisabled,
					},
					UpdateStatus:       types.CapacityProviderUpdateStatusDeleteComplete,
					UpdateStatusReason: adapters.PtrString("reason"),
				},
			},
		},
	}

	var page string

	if params.NextToken != nil {
		page = *params.NextToken
	}

	return pages[page], nil
}

func TestCapacityProviderOutputMapper(t *testing.T) {
	items, err := capacityProviderOutputMapper(
		context.Background(),
		&TestClient{},
		"foo",
		nil,
		&ecs.DescribeCapacityProvidersOutput{
			CapacityProviders: []types.CapacityProvider{
				{
					CapacityProviderArn: adapters.PtrString("arn:aws:ecs:eu-west-2:052392120703:capacity-provider/test"),
					Name:                adapters.PtrString("test"),
					Status:              types.CapacityProviderStatusActive,
					AutoScalingGroupProvider: &types.AutoScalingGroupProvider{
						AutoScalingGroupArn: adapters.PtrString("arn:aws:autoscaling:eu-west-2:052392120703:autoScalingGroup:9df90815-98c1-4136-a12a-90abef1c4e4e:autoScalingGroupName/ecs-test"),
						ManagedScaling: &types.ManagedScaling{
							Status:                 types.ManagedScalingStatusEnabled,
							TargetCapacity:         adapters.PtrInt32(80),
							MinimumScalingStepSize: adapters.PtrInt32(1),
							MaximumScalingStepSize: adapters.PtrInt32(10000),
							InstanceWarmupPeriod:   adapters.PtrInt32(300),
						},
						ManagedTerminationProtection: types.ManagedTerminationProtectionDisabled,
					},
					UpdateStatus:       types.CapacityProviderUpdateStatusDeleteComplete,
					UpdateStatusReason: adapters.PtrString("reason"),
				},
			},
		},
	)

	if err != nil {
		t.Error(err)
	}

	if len(items) != 1 {
		t.Errorf("expected 1 item, got %v", len(items))
	}

	item := items[0]

	if err = item.Validate(); err != nil {
		t.Error(err)
	}

	tests := adapters.QueryTests{
		{
			ExpectedType:   "autoscaling-auto-scaling-group",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "arn:aws:autoscaling:eu-west-2:052392120703:autoScalingGroup:9df90815-98c1-4136-a12a-90abef1c4e4e:autoScalingGroupName/ecs-test",
			ExpectedScope:  "052392120703.eu-west-2",
		},
	}

	tests.Execute(t, item)
}

func TestCapacityProviderAdapter(t *testing.T) {
	adapter := NewCapacityProviderAdapter(&TestClient{}, "", "")

	items, err := adapter.List(context.Background(), "", false)

	if err != nil {
		t.Error(err)
	}

	if len(items) != 3 {
		t.Errorf("expected 3 items, got %v", len(items))
	}
}

func TestNewCapacityProviderSource(t *testing.T) {
	config, account, region := adapters.GetAutoConfig(t)
	client := ecs.NewFromConfig(config)

	source := NewCapacityProviderAdapter(client, account, region)

	test := adapters.E2ETest{
		Adapter: source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}

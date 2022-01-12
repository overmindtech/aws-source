package elasticloadbalancing

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"testing"
	"time"

	elbv2 "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/overmindtech/discovery"
)

func TestNLBv2(t *testing.T) {
	t.Parallel()

	var err error
	elbClient := elbv2.NewFromConfig(TestAWSConfig)
	name := *TestVPC.ID + "test-elbv2"
	subnetIDs := make([]string, 0)

	for _, subnet := range TestVPC.Subnets {
		subnetIDs = append(subnetIDs, *subnet.ID)
	}

	var createLoadBalancerOutput *elbv2.CreateLoadBalancerOutput

	createLoadBalancerOutput, err = elbClient.CreateLoadBalancer(
		context.Background(),
		&elbv2.CreateLoadBalancerInput{
			Name:          &name,
			IpAddressType: types.IpAddressTypeIpv4,
			Scheme:        types.LoadBalancerSchemeEnumInternetFacing,
			Type:          types.LoadBalancerTypeEnumNetwork,
			Subnets:       subnetIDs,
		},
	)

	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_, err := elbClient.DeleteLoadBalancer(context.Background(), &elbv2.DeleteLoadBalancerInput{
			LoadBalancerArn: createLoadBalancerOutput.LoadBalancers[0].LoadBalancerArn,
		})

		if err != nil {
			t.Error(err)
		}

		// Wait in order to avoid race conditions where the load balancer hasn't
		// yet been fully deleted and subnet deletion fails
		time.Sleep(3 * time.Second)
	})

	var targetGroupOutput *elbv2.CreateTargetGroupOutput
	targetGroupName := "fake-targets"
	targetHealthCheckEnabled := true
	targetPort := int32(80)

	targetGroupOutput, err = elbClient.CreateTargetGroup(
		context.Background(),
		&elbv2.CreateTargetGroupInput{
			Name:                &targetGroupName,
			HealthCheckEnabled:  &targetHealthCheckEnabled,
			HealthCheckProtocol: types.ProtocolEnumTcp,
			Port:                &targetPort,
			TargetType:          types.TargetTypeEnumIp,
			Protocol:            types.ProtocolEnumTcp,
			VpcId:               TestVPC.ID,
		},
	)

	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_, err := elbClient.DeleteTargetGroup(
			context.Background(),
			&elbv2.DeleteTargetGroupInput{
				TargetGroupArn: targetGroupOutput.TargetGroups[0].TargetGroupArn,
			},
		)

		if err != nil {
			t.Error(err)
		}
	})

	cidrRegexp := regexp.MustCompile(`(\d+)\/\d+$`)

	for _, subnet := range TestVPC.Subnets {
		startNumber, _ := strconv.Atoi(cidrRegexp.FindStringSubmatch(subnet.CIDR)[1])

		targetIP := cidrRegexp.ReplaceAllString(subnet.CIDR, fmt.Sprint(startNumber+5))

		_, err = elbClient.RegisterTargets(
			context.Background(),
			&elbv2.RegisterTargetsInput{
				TargetGroupArn: targetGroupOutput.TargetGroups[0].TargetGroupArn,
				Targets: []types.TargetDescription{
					{
						Id:   &targetIP,
						Port: &targetPort,
					},
				},
			},
		)

		if err != nil {
			t.Fatal(err)
		}

		t.Cleanup(func() {
			_, err := elbClient.DeregisterTargets(
				context.Background(),
				&elbv2.DeregisterTargetsInput{
					TargetGroupArn: targetGroupOutput.TargetGroups[0].TargetGroupArn,
					Targets: []types.TargetDescription{
						{
							Id:   &targetIP,
							Port: &targetPort,
						},
					},
				},
			)

			if err != nil {
				t.Error(err)
			}
		})
	}

	var createListenerOutput *elbv2.CreateListenerOutput
	order := int32(1)

	createListenerOutput, err = elbClient.CreateListener(
		context.Background(),
		&elbv2.CreateListenerInput{
			LoadBalancerArn: createLoadBalancerOutput.LoadBalancers[0].LoadBalancerArn,
			Port:            &targetPort,
			Protocol:        types.ProtocolEnumTcp,
			DefaultActions: []types.Action{
				{
					Type:           types.ActionTypeEnumForward,
					TargetGroupArn: targetGroupOutput.TargetGroups[0].TargetGroupArn,
					Order:          &order,
				},
			},
		},
	)

	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_, err := elbClient.DeleteListener(
			context.Background(),
			&elbv2.DeleteListenerInput{
				ListenerArn: createListenerOutput.Listeners[0].ListenerArn,
			},
		)

		if err != nil {
			t.Error(err)
		}
	})

	elbWait := elbv2.NewLoadBalancerAvailableWaiter(
		elbClient,
	)

	err = elbWait.Wait(
		context.Background(),
		&elbv2.DescribeLoadBalancersInput{
			Names: []string{
				name,
			},
		},
		5*time.Minute,
	)

	if err != nil {
		t.Fatal(err)
	}

	src := ELBv2Source{
		Config:    TestAWSConfig,
		AccountID: TestAccountID,
	}

	t.Run("get NLB details", func(t *testing.T) {
		item, err := src.Get(context.Background(), TestContext, name)

		if err != nil {
			t.Fatal(err)
		}

		discovery.TestValidateItem(t, item)

		b, _ := json.Marshal(item)

		fmt.Println("---")
		fmt.Println(string(b))
		fmt.Println("---")
	})

	t.Run("get NLB that doesn't exist", func(t *testing.T) {
		_, err := src.Get(context.Background(), TestContext, "foobar")

		if err == nil {
			t.Error("expected error but got nil")
		}

	})

	t.Run("find all NLBs", func(t *testing.T) {
		items, err := src.Find(context.Background(), TestContext)

		if err != nil {
			t.Fatal(err)
		}

		if len(items) < 1 {
			t.Errorf("expected >=1 NLB but got %v", len(items))
		}
	})
}

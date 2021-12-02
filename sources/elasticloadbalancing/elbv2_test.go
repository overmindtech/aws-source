package elasticloadbalancing

import (
	"context"
	"fmt"
	"testing"
	"time"

	elbv2 "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/overmindtech/discovery"
)

func TestELBv2(t *testing.T) {
	t.Parallel()

	var err error
	elbClient := elbv2.NewFromConfig(TestAWSConfig)
	name := "test-elbv2"
	subnetIDs := make([]string, 0)

	for _, subnet := range TestVPC.Subnets {
		subnetIDs = append(subnetIDs, *subnet.ID)
	}

	var result *elbv2.CreateLoadBalancerOutput

	result, err = elbClient.CreateLoadBalancer(
		context.Background(),
		&elbv2.CreateLoadBalancerInput{
			Name:          &name,
			IpAddressType: types.IpAddressTypeIpv4,
			Scheme:        types.LoadBalancerSchemeEnumInternetFacing,
			Type:          types.LoadBalancerTypeEnumApplication,
			Subnets:       subnetIDs,
		},
	)

	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_, err := elbClient.DeleteLoadBalancer(context.Background(), &elbv2.DeleteLoadBalancerInput{
			LoadBalancerArn: result.LoadBalancers[0].LoadBalancerArn,
		})

		if err != nil {
			t.Error(err)
		}

		// Wait in order to avoid race conditions where the load balancer hasn't
		// yet been fully deleted and subnet deletion fails
		time.Sleep(3 * time.Second)
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
		3*time.Minute,
	)

	if err != nil {
		t.Fatal(err)
	}

	stsClient := sts.NewFromConfig(TestAWSConfig)

	var callerID *sts.GetCallerIdentityOutput

	callerID, err = stsClient.GetCallerIdentity(
		context.Background(),
		&sts.GetCallerIdentityInput{},
	)

	if err != nil {
		t.Fatal(err)
	}

	src := ELBv2Source{
		Config:    TestAWSConfig,
		AccountID: *callerID.Account,
	}

	testContext := fmt.Sprintf("%v.%v", *callerID.Account, TestAWSConfig.Region)

	t.Run("get elb details", func(t *testing.T) {
		item, err := src.Get(context.Background(), testContext, name)

		if err != nil {
			t.Fatal(err)
		}

		discovery.TestValidateItem(t, item)
	})

	t.Run("get elb that doesn't exist", func(t *testing.T) {
		_, err := src.Get(context.Background(), testContext, "foobar")

		if err == nil {
			t.Error("expected error but got nil")
		}

	})

	t.Run("find all ELBs", func(t *testing.T) {
		items, err := src.Find(context.Background(), testContext)

		if err != nil {
			t.Fatal(err)
		}

		if len(items) < 1 {
			t.Errorf("expected >=1 ELB but got %v", len(items))
		}
	})
}

package elasticloadbalancing

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	elb "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/overmindtech/discovery"
)

func TestELB(t *testing.T) {
	cfg, err := config.LoadDefaultConfig(context.TODO())

	if err != nil {
		t.Skipf("Config load failed: %v", err)
	}

	elbClient := elb.NewFromConfig(cfg)
	name := "test-elb"
	tag1key := "test-id"
	tag1value := "test"
	protocol := "TCP"

	_, err = elbClient.CreateLoadBalancer(
		context.Background(),
		&elb.CreateLoadBalancerInput{
			LoadBalancerName: &name,
			AvailabilityZones: []string{
				"eu-west-2a",
			},
			Listeners: []types.Listener{
				{
					InstancePort:     31572,
					LoadBalancerPort: 31572,
					Protocol:         &protocol,
					InstanceProtocol: &protocol,
				},
			},
			Tags: []types.Tag{
				{
					Key:   &tag1key,
					Value: &tag1value,
				},
			},
		},
	)

	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		elbClient.DeleteLoadBalancer(context.Background(), &elb.DeleteLoadBalancerInput{
			LoadBalancerName: &name,
		})
	})

	stsClient := sts.NewFromConfig(cfg)

	var callerID *sts.GetCallerIdentityOutput

	callerID, err = stsClient.GetCallerIdentity(
		context.Background(),
		&sts.GetCallerIdentityInput{},
	)

	if err != nil {
		t.Fatal(err)
	}

	src := ELBSource{
		Config:    cfg,
		AccountID: *callerID.Account,
	}

	testContext := fmt.Sprintf("%v.%v", callerID.Account, cfg.Region)

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

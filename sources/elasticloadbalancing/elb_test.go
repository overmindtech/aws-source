package elasticloadbalancing

import (
	"context"
	"testing"

	elb "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing/types"
	"github.com/overmindtech/discovery"
)

func TestELB(t *testing.T) {
	t.Parallel()

	var err error
	elbClient := elb.NewFromConfig(TestAWSConfig)
	name := *TestVPC.ID + "test-elb"
	tag1key := "test-id"
	tag1value := "test"
	protocol := "TCP"

	_, err = elbClient.CreateLoadBalancer(
		context.Background(),
		&elb.CreateLoadBalancerInput{
			LoadBalancerName: &name,
			AvailabilityZones: []string{
				TestVPC.Subnets[0].AvailabilityZone,
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

	src := ELBSource{
		Config:    TestAWSConfig,
		AccountID: TestAccountID,
	}

	t.Run("get elb details", func(t *testing.T) {
		item, err := src.Get(context.Background(), TestContext, name)

		if err != nil {
			t.Fatal(err)
		}

		discovery.TestValidateItem(t, item)
	})

	t.Run("get elb that doesn't exist", func(t *testing.T) {
		_, err := src.Get(context.Background(), TestContext, "foobar")

		if err == nil {
			t.Error("expected error but got nil")
		}

	})

	t.Run("find all ELBs", func(t *testing.T) {
		items, err := src.Find(context.Background(), TestContext)

		if err != nil {
			t.Fatal(err)
		}

		if len(items) < 1 {
			t.Errorf("expected >=1 ELB but got %v", len(items))
		}
	})
}

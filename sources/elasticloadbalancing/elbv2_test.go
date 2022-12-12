package elasticloadbalancing

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/overmindtech/aws-source/sources"
)

func TestELBv2Mapping(t *testing.T) {
	t.Parallel()

	t.Run("empty", func(t *testing.T) {
		lbName := "lbName"
		lb := ExpandedELBv2{
			LoadBalancer: types.LoadBalancer{
				LoadBalancerName: &lbName,
			},
		}

		item, err := mapExpandedELBv2ToItem(&lb, "foo.bar")
		if err != nil {
			t.Fatal(err)
		}
		if item == nil {
			t.Error("item is nil")
		}
	})
	t.Run("name check", func(t *testing.T) {
		lbName := ""
		lb := ExpandedELBv2{
			LoadBalancer: types.LoadBalancer{
				LoadBalancerName: &lbName,
			},
		}

		_, err := mapExpandedELBv2ToItem(&lb, "foo.bar")
		if err == nil {
			t.Fatal("didn't get expected error")
		}
	})
	t.Run("with DNSName", func(t *testing.T) {
		lbName := "lbName"
		dNSName := "dNSName"
		lb := ExpandedELBv2{
			LoadBalancer: types.LoadBalancer{
				LoadBalancerName: &lbName,
				DNSName:          &dNSName,
			},
		}

		item, err := mapExpandedELBv2ToItem(&lb, "foo.bar")
		if err != nil {
			t.Fatal(err)
		}
		if item == nil {
			t.Fatal("item is nil")
		}
		if len(item.LinkedItemRequests) != 1 {
			t.Fatalf("unexpected LinkedItemRequests: %v", item)
		}
		sources.CheckItem(t, item.LinkedItemRequests[0], "dNSName", "dns", dNSName, "global")
	})
}

package elasticloadbalancing

import (
	"context"
	"strings"
	"testing"

	elbv2 "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
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
		sources.CheckItemRequest(t, item.LinkedItemRequests[0], "dNSName", "dns", dNSName, "global")
	})
}

func TestGetv2(t *testing.T) {
	t.Parallel()
	t.Run("empty (scope mismatch)", func(t *testing.T) {
		src := ELBv2Source{}

		items, err := src.Get(context.Background(), "foo.bar", "query")
		if items != nil {
			t.Fatalf("unexpected items: %v", items)
		}
		if err == nil {
			t.Fatalf("expected err, got nil")
		}
		if !strings.HasPrefix(err.Error(), "requested scope foo.bar does not match source scope .") {
			t.Errorf("expected 'requested scope foo.bar does not match source scope .', got '%v'", err.Error())
		}
	})
}

type fakeV2Client struct {
	lbClientCalls           int
	listenerClientCalls     int
	targetGroupsClientCalls int

	DescribeLoadBalancersMock func(ctx context.Context, m fakeV2Client, params *elbv2.DescribeLoadBalancersInput, optFns ...func(*elbv2.Options)) (*elbv2.DescribeLoadBalancersOutput, error)
	DescribeListenersMock     func(ctx context.Context, m fakeV2Client, params *elbv2.DescribeListenersInput, optFns ...func(*elbv2.Options)) (*elbv2.DescribeListenersOutput, error)
	DescribeTargetGroupsMock  func(ctx context.Context, m fakeV2Client, params *elbv2.DescribeTargetGroupsInput, optFns ...func(*elbv2.Options)) (*elbv2.DescribeTargetGroupsOutput, error)
	DescribeTargetHealthMock  func(ctx context.Context, m fakeV2Client, params *elbv2.DescribeTargetHealthInput, optFns ...func(*elbv2.Options)) (*elbv2.DescribeTargetHealthOutput, error)
}

// DescribeListeners implements ELBv2Client
func (m fakeV2Client) DescribeListeners(ctx context.Context, params *elbv2.DescribeListenersInput, optFns ...func(*elbv2.Options)) (*elbv2.DescribeListenersOutput, error) {
	return m.DescribeListenersMock(ctx, m, params, optFns...)
}

// DescribeLoadBalancers implements ELBv2Client
func (m fakeV2Client) DescribeLoadBalancers(ctx context.Context, params *elbv2.DescribeLoadBalancersInput, optFns ...func(*elbv2.Options)) (*elbv2.DescribeLoadBalancersOutput, error) {
	return m.DescribeLoadBalancersMock(ctx, m, params, optFns...)
}

// DescribeTargetGroups implements ELBv2Client
func (m fakeV2Client) DescribeTargetGroups(ctx context.Context, params *elbv2.DescribeTargetGroupsInput, optFns ...func(*elbv2.Options)) (*elbv2.DescribeTargetGroupsOutput, error) {
	return m.DescribeTargetGroupsMock(ctx, m, params, optFns...)
}

// DescribeTargetHealth implements ELBv2Client
func (m fakeV2Client) DescribeTargetHealth(ctx context.Context, params *elbv2.DescribeTargetHealthInput, optFns ...func(*elbv2.Options)) (*elbv2.DescribeTargetHealthOutput, error) {
	return m.DescribeTargetHealthMock(ctx, m, params, optFns...)
}

func createFakeV2Client(t *testing.T) ELBv2Client {
	return fakeV2Client{
		DescribeLoadBalancersMock: func(ctx context.Context, m fakeV2Client, params *elbv2.DescribeLoadBalancersInput, optFns ...func(*elbv2.Options)) (*elbv2.DescribeLoadBalancersOutput, error) {
			m.lbClientCalls += 1
			if m.lbClientCalls > 2 {
				t.Error("Called DescribeLoadBalancersMock too often (>2)")
				return nil, nil
			}
			if params.Marker == nil {
				firstName := "first"
				nextMarker := "page2"
				return &elbv2.DescribeLoadBalancersOutput{
					LoadBalancers: []types.LoadBalancer{
						{
							LoadBalancerName: &firstName,
						},
					},
					NextMarker: &nextMarker,
				}, nil
			} else if *params.Marker == "page2" {
				secondName := "second"
				return &elbv2.DescribeLoadBalancersOutput{
					LoadBalancers: []types.LoadBalancer{
						{
							LoadBalancerName: &secondName,
						},
					},
				}, nil
			}
			return nil, nil
		},
		DescribeListenersMock: func(ctx context.Context, m fakeV2Client, params *elbv2.DescribeListenersInput, optFns ...func(*elbv2.Options)) (*elbv2.DescribeListenersOutput, error) {
			m.listenerClientCalls += 1
			if m.listenerClientCalls > 2 {
				t.Error("Called DescribeListenersMock too often (>2)")
				return nil, nil
			}
			if params.Marker == nil {
				// firstName := "first"
				nextMarker := "page2"
				return &elbv2.DescribeListenersOutput{
					Listeners: []types.Listener{
						{},
					},
					NextMarker: &nextMarker,
				}, nil
			} else if *params.Marker == "page2" {
				// secondName := "second"
				return &elbv2.DescribeListenersOutput{
					Listeners: []types.Listener{
						{},
					},
				}, nil
			}
			return nil, nil
		},
		DescribeTargetGroupsMock: func(ctx context.Context, m fakeV2Client, params *elbv2.DescribeTargetGroupsInput, optFns ...func(*elbv2.Options)) (*elbv2.DescribeTargetGroupsOutput, error) {
			m.targetGroupsClientCalls += 1
			if m.targetGroupsClientCalls > 2 {
				t.Error("Called DescribeTargetGroupsMock too often (>2)")
				return nil, nil
			}
			if params.Marker == nil {
				// firstName := "first"
				nextMarker := "page2"
				return &elbv2.DescribeTargetGroupsOutput{
					NextMarker: &nextMarker,
				}, nil
			} else if *params.Marker == "page2" {
				// secondName := "second"
				return &elbv2.DescribeTargetGroupsOutput{}, nil
			}
			return nil, nil
		},
		DescribeTargetHealthMock: func(ctx context.Context, m fakeV2Client, params *elbv2.DescribeTargetHealthInput, optFns ...func(*elbv2.Options)) (*elbv2.DescribeTargetHealthOutput, error) {
			return nil, nil
		},
	}
}

func TestGetV2Impl(t *testing.T) {
	t.Parallel()
	t.Run("with client", func(t *testing.T) {
		item, err := getv2Impl(context.Background(), createFakeV2Client(t), "foo.bar", "query")
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if item == nil {
			t.Fatalf("item is nil")
		}
		if item.Attributes.AttrStruct.Fields["name"].GetStringValue() != "first" {
			t.Errorf("unexpected first item: %v", item)
		}
	})
}

func TestListV2(t *testing.T) {
	t.Parallel()
	t.Run("empty (scope mismatch)", func(t *testing.T) {
		src := ELBv2Source{}

		items, err := src.List(context.Background(), "foo.bar")
		if items != nil {
			t.Fatalf("unexpected items: %v", items)
		}
		if err == nil {
			t.Fatalf("expected err, got nil")
		}
		if !strings.HasPrefix(err.Error(), "requested scope foo.bar does not match source scope .") {
			t.Errorf("expected 'requested scope foo.bar does not match source scope .', got '%v'", err.Error())
		}
	})
}

func TestListV2Impl(t *testing.T) {
	t.Parallel()
	t.Run("with client", func(t *testing.T) {
		items, err := findV2Impl(context.Background(), createFakeV2Client(t), "foo.bar")
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if len(items) != 2 {
			t.Fatalf("unexpected items (len=%v): %v", len(items), items)
		}
		if items[0].Attributes.AttrStruct.Fields["name"].GetStringValue() != "first" {
			t.Errorf("unexpected first item: %v", items[0])
		}
		if items[1].Attributes.AttrStruct.Fields["name"].GetStringValue() != "second" {
			t.Errorf("unexpected second item: %v", items[0])
		}
	})
}

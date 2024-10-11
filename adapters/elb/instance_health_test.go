package elb

import (
	"context"
	"testing"

	elb "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing/types"
	"github.com/overmindtech/aws-source/adapters"
	"github.com/overmindtech/sdp-go"
)

func TestInstanceHealthOutputMapper(t *testing.T) {

	output := elb.DescribeInstanceHealthOutput{
		InstanceStates: []types.InstanceState{
			{
				InstanceId:  adapters.PtrString("i-0337802d908b4a81e"), // link
				State:       adapters.PtrString("InService"),
				ReasonCode:  adapters.PtrString("N/A"),
				Description: adapters.PtrString("N/A"),
			},
		},
	}

	items, err := instanceHealthOutputMapper(context.Background(), nil, "foo", nil, &output)

	if err != nil {
		t.Error(err)
	}

	for _, item := range items {
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
			ExpectedType:   "ec2-instance",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "i-0337802d908b4a81e",
			ExpectedScope:  "foo",
		},
	}

	tests.Execute(t, item)
}

package ec2

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func TestInstanceStatusInputMapperGet(t *testing.T) {
	input, err := InstanceStatusInputMapperGet("foo", "bar")

	if err != nil {
		t.Error(err)
	}

	if len(input.InstanceIds) != 1 {
		t.Fatalf("expected 1 instanceStatus ID, got %v", len(input.InstanceIds))
	}

	if input.InstanceIds[0] != "bar" {
		t.Errorf("expected instanceStatus ID to be bar, got %v", input.InstanceIds[0])
	}
}

func TestInstanceStatusInputMapperList(t *testing.T) {
	input, err := InstanceStatusInputMapperList("foo")

	if err != nil {
		t.Error(err)
	}

	if len(input.Filters) != 0 || len(input.InstanceIds) != 0 {
		t.Errorf("non-empty input: %v", input)
	}
}

func TestInstanceStatusOutputMapper(t *testing.T) {
	output := &ec2.DescribeInstanceStatusOutput{
		InstanceStatuses: []types.InstanceStatus{
			{
				AvailabilityZone: sources.PtrString("eu-west-2c"),
				InstanceId:       sources.PtrString("i-022bdccde30270570"),
				InstanceState: &types.InstanceState{
					Code: sources.PtrInt32(16),
					Name: types.InstanceStateNameRunning,
				},
				InstanceStatus: &types.InstanceStatusSummary{
					Details: []types.InstanceStatusDetails{
						{
							Name:   types.StatusNameReachability,
							Status: types.StatusTypePassed,
						},
					},
					Status: types.SummaryStatusOk,
				},
				SystemStatus: &types.InstanceStatusSummary{
					Details: []types.InstanceStatusDetails{
						{
							Name:   types.StatusNameReachability,
							Status: types.StatusTypePassed,
						},
					},
					Status: types.SummaryStatusOk,
				},
			},
		},
	}

	items, err := InstanceStatusOutputMapper("foo", output)

	if err != nil {
		t.Fatal(err)
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
	tests := sources.ItemRequestTests{
		{
			ExpectedType:   "ec2-availability-zone",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  "eu-west-2c",
			ExpectedScope:  item.Scope,
		},
	}

	tests.Execute(t, item)

}

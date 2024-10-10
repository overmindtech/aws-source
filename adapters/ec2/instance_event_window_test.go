package ec2

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/overmindtech/aws-source/adapters"
	"github.com/overmindtech/sdp-go"
)

func TestInstanceEventWindowInputMapperGet(t *testing.T) {
	input, err := instanceEventWindowInputMapperGet("foo", "bar")

	if err != nil {
		t.Error(err)
	}

	if len(input.InstanceEventWindowIds) != 1 {
		t.Fatalf("expected 1 InstanceEventWindow ID, got %v", len(input.InstanceEventWindowIds))
	}

	if input.InstanceEventWindowIds[0] != "bar" {
		t.Errorf("expected InstanceEventWindow ID to be bar, got %v", input.InstanceEventWindowIds[0])
	}
}

func TestInstanceEventWindowInputMapperList(t *testing.T) {
	input, err := instanceEventWindowInputMapperList("foo")

	if err != nil {
		t.Error(err)
	}

	if len(input.Filters) != 0 || len(input.InstanceEventWindowIds) != 0 {
		t.Errorf("non-empty input: %v", input)
	}
}

func TestInstanceEventWindowOutputMapper(t *testing.T) {
	output := &ec2.DescribeInstanceEventWindowsOutput{
		InstanceEventWindows: []types.InstanceEventWindow{
			{
				AssociationTarget: &types.InstanceEventWindowAssociationTarget{
					DedicatedHostIds: []string{
						"dedicated",
					},
					InstanceIds: []string{
						"instance",
					},
				},
				CronExpression:        adapters.PtrString("something"),
				InstanceEventWindowId: adapters.PtrString("window-123"),
				Name:                  adapters.PtrString("test"),
				State:                 types.InstanceEventWindowStateActive,
				TimeRanges: []types.InstanceEventWindowTimeRange{
					{
						StartHour:    adapters.PtrInt32(1),
						EndHour:      adapters.PtrInt32(2),
						EndWeekDay:   types.WeekDayFriday,
						StartWeekDay: types.WeekDayMonday,
					},
				},
				Tags: []types.Tag{},
			},
		},
	}

	items, err := instanceEventWindowOutputMapper(context.Background(), nil, "foo", nil, output)

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
	tests := adapters.QueryTests{
		{
			ExpectedType:   "ec2-host",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "dedicated",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "ec2-instance",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "instance",
			ExpectedScope:  "foo",
		},
	}

	tests.Execute(t, item)

}

func TestNewInstanceEventWindowSource(t *testing.T) {
	client, account, region := GetAutoConfig(t)

	source := NewInstanceEventWindowSource(client, account, region)

	test := adapters.E2ETest{
		Adapter: source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}

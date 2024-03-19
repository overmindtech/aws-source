package ec2

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/overmindtech/aws-source/sources"
)

func TestPlacementGroupInputMapperGet(t *testing.T) {
	input, err := placementGroupInputMapperGet("foo", "bar")

	if err != nil {
		t.Error(err)
	}

	if len(input.GroupIds) != 1 {
		t.Fatalf("expected 1 PlacementGroup ID, got %v", len(input.GroupIds))
	}

	if input.GroupIds[0] != "bar" {
		t.Errorf("expected PlacementGroup ID to be bar, got %v", input.GroupIds[0])
	}
}

func TestPlacementGroupInputMapperList(t *testing.T) {
	input, err := placementGroupInputMapperList("foo")

	if err != nil {
		t.Error(err)
	}

	if len(input.Filters) != 0 || len(input.GroupIds) != 0 {
		t.Errorf("non-empty input: %v", input)
	}
}

func TestPlacementGroupOutputMapper(t *testing.T) {
	output := &ec2.DescribePlacementGroupsOutput{
		PlacementGroups: []types.PlacementGroup{
			{
				GroupArn:       sources.PtrString("arn"),
				GroupId:        sources.PtrString("id"),
				GroupName:      sources.PtrString("name"),
				SpreadLevel:    types.SpreadLevelHost,
				State:          types.PlacementGroupStateAvailable,
				Strategy:       types.PlacementStrategyCluster,
				PartitionCount: sources.PtrInt32(1),
				Tags:           []types.Tag{},
			},
		},
	}

	items, err := placementGroupOutputMapper(context.Background(), nil, "foo", nil, output)

	if err != nil {
		t.Fatal(err)
	}

	for _, item := range items {
		if err := item.Validate(); err != nil {
			t.Error(err)
		}
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 items, got %v", len(items))
	}

}

func TestNewPlacementGroupSource(t *testing.T) {
	client, account, region := GetAutoConfig(t)

	source := NewPlacementGroupSource(client, account, region, &TestRateLimit)

	test := sources.E2ETest{
		Source:  source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}

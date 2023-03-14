package ec2

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/overmindtech/aws-source/sources"
)

func TestRegionInputMapperGet(t *testing.T) {
	input, err := regionInputMapperGet("foo", "bar")

	if err != nil {
		t.Error(err)
	}

	if len(input.RegionNames) != 1 {
		t.Fatalf("expected 1 Region ID, got %v", len(input.RegionNames))
	}

	if input.RegionNames[0] != "bar" {
		t.Errorf("expected Region ID to be bar, got %v", input.RegionNames[0])
	}
}

func TestRegionInputMapperList(t *testing.T) {
	input, err := regionInputMapperList("foo")

	if err != nil {
		t.Error(err)
	}

	if len(input.Filters) != 0 || len(input.RegionNames) != 0 {
		t.Errorf("non-empty input: %v", input)
	}
}

func TestRegionOutputMapper(t *testing.T) {
	output := &ec2.DescribeRegionsOutput{
		Regions: []types.Region{
			{
				Endpoint:    sources.PtrString("ec2.ap-south-1.amazonaws.com"),
				RegionName:  sources.PtrString("ap-south-1"),
				OptInStatus: sources.PtrString("opt-in-not-required"),
			},
			{
				Endpoint:    sources.PtrString("ec2.eu-north-1.amazonaws.com"),
				RegionName:  sources.PtrString("eu-north-1"),
				OptInStatus: sources.PtrString("opt-in-not-required"),
			},
			{
				Endpoint:    sources.PtrString("ec2.eu-west-3.amazonaws.com"),
				RegionName:  sources.PtrString("eu-west-3"),
				OptInStatus: sources.PtrString("opt-in-not-required"),
			},
			{
				Endpoint:    sources.PtrString("ec2.eu-west-2.amazonaws.com"),
				RegionName:  sources.PtrString("eu-west-2"),
				OptInStatus: sources.PtrString("opt-in-not-required"),
			},
		},
	}

	items, err := regionOutputMapper("foo", nil, output)

	if err != nil {
		t.Fatal(err)
	}

	for _, item := range items {
		if err := item.Validate(); err != nil {
			t.Error(err)
		}
	}

	if len(items) != 4 {
		t.Fatalf("expected 4 items, got %v", len(items))
	}

}

func TestNewRegionSource(t *testing.T) {
	config, account, _ := sources.GetAutoConfig(t)

	rateLimit := LimitBucket{
		MaxCapacity: 50,
		RefillRate:  10,
	}

	rateLimitCtx, rateLimitCancel := context.WithCancel(context.Background())
	defer rateLimitCancel()

	rateLimit.Start(rateLimitCtx)

	source := NewRegionSource(config, account, &rateLimit)

	test := sources.E2ETest{
		Source:  source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}

package ec2

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/overmindtech/aws-source/sources"
)

func TestLaunchTemplateInputMapperGet(t *testing.T) {
	input, err := LaunchTemplateInputMapperGet("foo", "bar")

	if err != nil {
		t.Error(err)
	}

	if len(input.LaunchTemplateIds) != 1 {
		t.Fatalf("expected 1 LaunchTemplate ID, got %v", len(input.LaunchTemplateIds))
	}

	if input.LaunchTemplateIds[0] != "bar" {
		t.Errorf("expected LaunchTemplate ID to be bar, got %v", input.LaunchTemplateIds[0])
	}
}

func TestLaunchTemplateInputMapperList(t *testing.T) {
	input, err := LaunchTemplateInputMapperList("foo")

	if err != nil {
		t.Error(err)
	}

	if len(input.Filters) != 0 || len(input.LaunchTemplateIds) != 0 {
		t.Errorf("non-empty input: %v", input)
	}
}

func TestLaunchTemplateOutputMapper(t *testing.T) {
	output := &ec2.DescribeLaunchTemplatesOutput{
		LaunchTemplates: []types.LaunchTemplate{
			{
				CreateTime:           sources.PtrTime(time.Now()),
				CreatedBy:            sources.PtrString("me"),
				DefaultVersionNumber: sources.PtrInt64(1),
				LatestVersionNumber:  sources.PtrInt64(10),
				LaunchTemplateId:     sources.PtrString("id"),
				LaunchTemplateName:   sources.PtrString("hello"),
				Tags:                 []types.Tag{},
			},
		},
	}

	items, err := LaunchTemplateOutputMapper("foo", nil, output)

	if err != nil {
		t.Fatal(err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %v", len(items))
	}

}

func TestNewLaunchTemplateSource(t *testing.T) {
	config, account, _ := sources.GetAutoConfig(t)

	rateLimit := LimitBucket{
		MaxCapacity: 50,
		RefillRate:  10,
	}

	rateLimitCtx, rateLimitCancel := context.WithCancel(context.Background())
	defer rateLimitCancel()

	rateLimit.Start(rateLimitCtx)

	source := NewLaunchTemplateSource(config, account, &rateLimit)

	test := sources.E2ETest{
		Source:  source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}

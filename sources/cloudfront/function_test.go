package cloudfront

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/overmindtech/aws-source/sources"
)

func TestFunctionItemMapper(t *testing.T) {
	summary := types.FunctionSummary{
		FunctionConfig: &types.FunctionConfig{
			Comment: sources.PtrString("test-comment"),
			Runtime: types.FunctionRuntimeCloudfrontJs20,
		},
		FunctionMetadata: &types.FunctionMetadata{
			FunctionARN:      sources.PtrString("arn:aws:cloudfront::123456789012:function/test-function"),
			LastModifiedTime: sources.PtrTime(time.Now()),
			CreatedTime:      sources.PtrTime(time.Now()),
			Stage:            types.FunctionStageLive,
		},
		Name:   sources.PtrString("test-function"),
		Status: sources.PtrString("test-status"),
	}

	item, err := functionItemMapper("test", &summary)

	if err != nil {
		t.Fatal(err)
	}

	if err = item.Validate(); err != nil {
		t.Error(err)
	}
}

func TestNewFunctionSource(t *testing.T) {
	config, account, _ := sources.GetAutoConfig(t)

	source := NewFunctionSource(config, account)

	test := sources.E2ETest{
		Source:  source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}

package cloudfront

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/overmindtech/aws-source/adapters"
)

func TestFunctionItemMapper(t *testing.T) {
	summary := types.FunctionSummary{
		FunctionConfig: &types.FunctionConfig{
			Comment: adapters.PtrString("test-comment"),
			Runtime: types.FunctionRuntimeCloudfrontJs20,
		},
		FunctionMetadata: &types.FunctionMetadata{
			FunctionARN:      adapters.PtrString("arn:aws:cloudfront::123456789012:function/test-function"),
			LastModifiedTime: adapters.PtrTime(time.Now()),
			CreatedTime:      adapters.PtrTime(time.Now()),
			Stage:            types.FunctionStageLive,
		},
		Name:   adapters.PtrString("test-function"),
		Status: adapters.PtrString("test-status"),
	}

	item, err := functionItemMapper("", "test", &summary)

	if err != nil {
		t.Fatal(err)
	}

	if err = item.Validate(); err != nil {
		t.Error(err)
	}
}

func TestNewFunctionAdapter(t *testing.T) {
	client, account, _ := GetAutoConfig(t)

	adapter := NewFunctionAdapter(client, account)

	test := adapters.E2ETest{
		Adapter: adapter,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}

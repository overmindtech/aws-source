package cloudfront

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/overmindtech/aws-source/adapters"
	"github.com/overmindtech/sdp-go"
)

func TestRealtimeLogConfigsItemMapper(t *testing.T) {
	x := types.RealtimeLogConfig{
		Name:         adapters.PtrString("test"),
		SamplingRate: adapters.PtrInt64(100),
		ARN:          adapters.PtrString("arn:aws:cloudfront::123456789012:realtime-log-config/12345678-1234-1234-1234-123456789012"),
		EndPoints: []types.EndPoint{
			{
				StreamType: adapters.PtrString("Kinesis"),
				KinesisStreamConfig: &types.KinesisStreamConfig{
					RoleARN:   adapters.PtrString("arn:aws:iam::123456789012:role/CloudFront_Logger"),              // link
					StreamARN: adapters.PtrString("arn:aws:kinesis:us-east-1:123456789012:stream/cloudfront-logs"), // link
				},
			},
		},
		Fields: []string{
			"date",
		},
	}

	item, err := realtimeLogConfigsItemMapper("", "test", &x)

	if err != nil {
		t.Fatal(err)
	}

	if err = item.Validate(); err != nil {
		t.Error(err)
	}

	tests := adapters.QueryTests{
		{
			ExpectedType:   "iam-role",
			ExpectedQuery:  "arn:aws:iam::123456789012:role/CloudFront_Logger",
			ExpectedScope:  "123456789012",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
		},
		{
			ExpectedType:   "kinesis-stream",
			ExpectedQuery:  "arn:aws:kinesis:us-east-1:123456789012:stream/cloudfront-logs",
			ExpectedScope:  "123456789012.us-east-1",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
		},
	}

	tests.Execute(t, item)
}

func TestNewRealtimeLogConfigsSource(t *testing.T) {
	client, account, _ := GetAutoConfig(t)

	source := NewRealtimeLogConfigsSource(client, account)

	test := adapters.E2ETest{
		Adapter: source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}

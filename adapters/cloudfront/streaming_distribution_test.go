package cloudfront

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/overmindtech/aws-source/adapters"
	"github.com/overmindtech/sdp-go"
)

func (t TestCloudFrontClient) GetStreamingDistribution(ctx context.Context, params *cloudfront.GetStreamingDistributionInput, optFns ...func(*cloudfront.Options)) (*cloudfront.GetStreamingDistributionOutput, error) {
	return &cloudfront.GetStreamingDistributionOutput{
		ETag: adapters.PtrString("E2QWRUHAPOMQZL"),
		StreamingDistribution: &types.StreamingDistribution{
			ARN:              adapters.PtrString("arn:aws:cloudfront::123456789012:streaming-distribution/EDFDVBD632BHDS5"),
			DomainName:       adapters.PtrString("d111111abcdef8.cloudfront.net"), // link
			Id:               adapters.PtrString("EDFDVBD632BHDS5"),
			Status:           adapters.PtrString("Deployed"), // health
			LastModifiedTime: adapters.PtrTime(time.Now()),
			ActiveTrustedSigners: &types.ActiveTrustedSigners{
				Enabled:  adapters.PtrBool(true),
				Quantity: adapters.PtrInt32(1),
				Items: []types.Signer{
					{
						AwsAccountNumber: adapters.PtrString("123456789012"),
						KeyPairIds: &types.KeyPairIds{
							Quantity: adapters.PtrInt32(1),
							Items: []string{
								"APKAJDGKZRVEXAMPLE",
							},
						},
					},
				},
			},
			StreamingDistributionConfig: &types.StreamingDistributionConfig{
				CallerReference: adapters.PtrString("test"),
				Comment:         adapters.PtrString("test"),
				Enabled:         adapters.PtrBool(true),
				S3Origin: &types.S3Origin{
					DomainName:           adapters.PtrString("myawsbucket.s3.amazonaws.com"),                     // link
					OriginAccessIdentity: adapters.PtrString("origin-access-identity/cloudfront/E127EXAMPLE51Z"), // link
				},
				TrustedSigners: &types.TrustedSigners{
					Enabled:  adapters.PtrBool(true),
					Quantity: adapters.PtrInt32(1),
					Items: []string{
						"self",
					},
				},
				Aliases: &types.Aliases{
					Quantity: adapters.PtrInt32(1),
					Items: []string{
						"example.com", // link
					},
				},
				Logging: &types.StreamingLoggingConfig{
					Bucket:  adapters.PtrString("myawslogbucket.s3.amazonaws.com"), // link
					Enabled: adapters.PtrBool(true),
					Prefix:  adapters.PtrString("myprefix"),
				},
				PriceClass: types.PriceClassPriceClassAll,
			},
		},
	}, nil
}

func (t TestCloudFrontClient) ListStreamingDistributions(ctx context.Context, params *cloudfront.ListStreamingDistributionsInput, optFns ...func(*cloudfront.Options)) (*cloudfront.ListStreamingDistributionsOutput, error) {
	return &cloudfront.ListStreamingDistributionsOutput{
		StreamingDistributionList: &types.StreamingDistributionList{
			IsTruncated: adapters.PtrBool(false),
			Items: []types.StreamingDistributionSummary{
				{
					Id: adapters.PtrString("test-id"),
				},
			},
		},
	}, nil
}

func TestStreamingDistributionGetFunc(t *testing.T) {
	item, err := streamingDistributionGetFunc(context.Background(), TestCloudFrontClient{}, "foo", &cloudfront.GetStreamingDistributionInput{})

	if err != nil {
		t.Fatal(err)
	}

	if err = item.Validate(); err != nil {
		t.Error(err)
	}

	if item.GetHealth() != sdp.Health_HEALTH_OK {
		t.Errorf("expected health to be HEALTH_OK, got %s", item.GetHealth())
	}

	tests := adapters.QueryTests{
		{
			ExpectedType:   "dns",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "d111111abcdef8.cloudfront.net",
			ExpectedScope:  "global",
		},
	}

	tests.Execute(t, item)
}

func TestNewStreamingDistributionAdapter(t *testing.T) {
	config, account, _ := adapters.GetAutoConfig(t)
	client := cloudfront.NewFromConfig(config)

	adapter := NewStreamingDistributionAdapter(client, account)

	test := adapters.E2ETest{
		Adapter: adapter,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}

package cloudfront

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func (t TestCloudFrontClient) GetStreamingDistribution(ctx context.Context, params *cloudfront.GetStreamingDistributionInput, optFns ...func(*cloudfront.Options)) (*cloudfront.GetStreamingDistributionOutput, error) {
	return &cloudfront.GetStreamingDistributionOutput{
		ETag: sources.PtrString("E2QWRUHAPOMQZL"),
		StreamingDistribution: &types.StreamingDistribution{
			ARN:              sources.PtrString("arn:aws:cloudfront::123456789012:streaming-distribution/EDFDVBD632BHDS5"),
			DomainName:       sources.PtrString("d111111abcdef8.cloudfront.net"), // link
			Id:               sources.PtrString("EDFDVBD632BHDS5"),
			Status:           sources.PtrString("Deployed"), // health
			LastModifiedTime: sources.PtrTime(time.Now()),
			ActiveTrustedSigners: &types.ActiveTrustedSigners{
				Enabled:  sources.PtrBool(true),
				Quantity: sources.PtrInt32(1),
				Items: []types.Signer{
					{
						AwsAccountNumber: sources.PtrString("123456789012"),
						KeyPairIds: &types.KeyPairIds{
							Quantity: sources.PtrInt32(1),
							Items: []string{
								"APKAJDGKZRVEXAMPLE",
							},
						},
					},
				},
			},
			StreamingDistributionConfig: &types.StreamingDistributionConfig{
				CallerReference: sources.PtrString("test"),
				Comment:         sources.PtrString("test"),
				Enabled:         sources.PtrBool(true),
				S3Origin: &types.S3Origin{
					DomainName:           sources.PtrString("myawsbucket.s3.amazonaws.com"),                     // link
					OriginAccessIdentity: sources.PtrString("origin-access-identity/cloudfront/E127EXAMPLE51Z"), // link
				},
				TrustedSigners: &types.TrustedSigners{
					Enabled:  sources.PtrBool(true),
					Quantity: sources.PtrInt32(1),
					Items: []string{
						"self",
					},
				},
				Aliases: &types.Aliases{
					Quantity: sources.PtrInt32(1),
					Items: []string{
						"example.com", // link
					},
				},
				Logging: &types.StreamingLoggingConfig{
					Bucket:  sources.PtrString("myawslogbucket.s3.amazonaws.com"), // link
					Enabled: sources.PtrBool(true),
					Prefix:  sources.PtrString("myprefix"),
				},
				PriceClass: types.PriceClassPriceClassAll,
			},
		},
	}, nil
}

func (t TestCloudFrontClient) ListStreamingDistributions(ctx context.Context, params *cloudfront.ListStreamingDistributionsInput, optFns ...func(*cloudfront.Options)) (*cloudfront.ListStreamingDistributionsOutput, error) {
	return &cloudfront.ListStreamingDistributionsOutput{
		StreamingDistributionList: &types.StreamingDistributionList{
			IsTruncated: sources.PtrBool(false),
			Items: []types.StreamingDistributionSummary{
				{
					Id: sources.PtrString("test-id"),
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

	tests := sources.QueryTests{
		{
			ExpectedType:   "dns",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "d111111abcdef8.cloudfront.net",
			ExpectedScope:  "global",
		},
	}

	tests.Execute(t, item)
}

func TestNewStreamingDistributionSource(t *testing.T) {
	config, account, _ := sources.GetAutoConfig(t)
	client := cloudfront.NewFromConfig(config)

	source := NewStreamingDistributionSource(client, account)

	test := sources.E2ETest{
		Source:  source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}

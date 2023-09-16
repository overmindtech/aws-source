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

func (t TestCloudFrontClient) GetDistribution(ctx context.Context, params *cloudfront.GetDistributionInput, optFns ...func(*cloudfront.Options)) (*cloudfront.GetDistributionOutput, error) {
	return &cloudfront.GetDistributionOutput{
		Distribution: &types.Distribution{
			ARN:                           sources.PtrString("arn:aws:cloudfront::123456789012:distribution/test-id"),
			DomainName:                    sources.PtrString("d111111abcdef8.cloudfront.net"), // link
			Id:                            sources.PtrString("test-id"),
			InProgressInvalidationBatches: sources.PtrInt32(1),
			LastModifiedTime:              sources.PtrTime(time.Now()),
			Status:                        sources.PtrString("Deployed"), // health: https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/distribution-web-values-returned.html
			ActiveTrustedKeyGroups: &types.ActiveTrustedKeyGroups{
				Enabled:  sources.PtrBool(true),
				Quantity: sources.PtrInt32(1),
				Items: []types.KGKeyPairIds{
					{
						KeyGroupId: sources.PtrString("key-group-1"), // link
						KeyPairIds: &types.KeyPairIds{
							Quantity: sources.PtrInt32(1),
							Items: []string{
								"123456789",
							},
						},
					},
				},
			},
			ActiveTrustedSigners: &types.ActiveTrustedSigners{
				Enabled:  sources.PtrBool(true),
				Quantity: sources.PtrInt32(1),
				Items: []types.Signer{
					{
						AwsAccountNumber: sources.PtrString("123456789"),
						KeyPairIds: &types.KeyPairIds{
							Quantity: sources.PtrInt32(1),
							Items: []string{
								"123456789",
							},
						},
					},
				},
			},
			AliasICPRecordals: []types.AliasICPRecordal{
				{
					CNAME:             sources.PtrString("something.foo.bar.com"), // link
					ICPRecordalStatus: types.ICPRecordalStatusApproved,
				},
			},
			DistributionConfig: &types.DistributionConfig{
				CallerReference: sources.PtrString("test-caller-reference"),
				Comment:         sources.PtrString("test-comment"),
				Enabled:         sources.PtrBool(true),
				Aliases: &types.Aliases{
					Quantity: sources.PtrInt32(1),
					Items: []string{
						"www.example.com", // link
					},
				},
				Staging:                      sources.PtrBool(true),
				ContinuousDeploymentPolicyId: sources.PtrString("test-continuous-deployment-policy-id"), // link
				CacheBehaviors: &types.CacheBehaviors{
					Quantity: sources.PtrInt32(1),
					Items: []types.CacheBehavior{
						{
							PathPattern:          sources.PtrString("/foo"),
							TargetOriginId:       sources.PtrString("CustomOriginConfig"),
							ViewerProtocolPolicy: types.ViewerProtocolPolicyHttpsOnly,
							AllowedMethods: &types.AllowedMethods{
								Items: []types.Method{
									types.MethodGet,
								},
							},
							CachePolicyId:           sources.PtrString("test-cache-policy-id"), // link
							Compress:                sources.PtrBool(true),
							DefaultTTL:              sources.PtrInt64(1),
							FieldLevelEncryptionId:  sources.PtrString("test-field-level-encryption-id"), // link
							MaxTTL:                  sources.PtrInt64(1),
							MinTTL:                  sources.PtrInt64(1),
							OriginRequestPolicyId:   sources.PtrString("test-origin-request-policy-id"),                                   // link
							RealtimeLogConfigArn:    sources.PtrString("arn:aws:logs:us-east-1:123456789012:realtime-log-config/test-id"), // link
							ResponseHeadersPolicyId: sources.PtrString("test-response-headers-policy-id"),                                 // link
							SmoothStreaming:         sources.PtrBool(true),
							TrustedKeyGroups: &types.TrustedKeyGroups{
								Enabled:  sources.PtrBool(true),
								Quantity: sources.PtrInt32(1),
								Items: []string{
									"key-group-1", // link
								},
							},
							TrustedSigners: &types.TrustedSigners{
								Enabled:  sources.PtrBool(true),
								Quantity: sources.PtrInt32(1),
								Items: []string{
									"123456789",
								},
							},
							ForwardedValues: &types.ForwardedValues{
								Cookies: &types.CookiePreference{
									Forward: types.ItemSelectionWhitelist,
									WhitelistedNames: &types.CookieNames{
										Quantity: sources.PtrInt32(1),
										Items: []string{
											"cookie_123",
										},
									},
								},
								QueryString: sources.PtrBool(true),
								Headers: &types.Headers{
									Quantity: sources.PtrInt32(1),
									Items: []string{
										"X-Customer-Header",
									},
								},
								QueryStringCacheKeys: &types.QueryStringCacheKeys{
									Quantity: sources.PtrInt32(1),
									Items: []string{
										"test-query-string-cache-key",
									},
								},
							},
							FunctionAssociations: &types.FunctionAssociations{
								Quantity: sources.PtrInt32(1),
								Items: []types.FunctionAssociation{
									{
										EventType:   types.EventTypeOriginRequest,
										FunctionARN: sources.PtrString("arn:aws:cloudfront::123412341234:function/1234"), // link
									},
								},
							},
							LambdaFunctionAssociations: &types.LambdaFunctionAssociations{
								Quantity: sources.PtrInt32(1),
								Items: []types.LambdaFunctionAssociation{
									{
										EventType:         types.EventTypeOriginResponse,
										LambdaFunctionARN: sources.PtrString("arn:aws:lambda:us-east-1:123456789012:function:test-function"), // link
										IncludeBody:       sources.PtrBool(true),
									},
								},
							},
						},
					},
				},
				Origins: &types.Origins{
					Items: []types.Origin{
						{
							DomainName:         sources.PtrString("DOC-EXAMPLE-BUCKET.s3.us-west-2.amazonaws.com"), // link
							Id:                 sources.PtrString("CustomOriginConfig"),
							ConnectionAttempts: sources.PtrInt32(3),
							ConnectionTimeout:  sources.PtrInt32(10),
							CustomHeaders: &types.CustomHeaders{
								Quantity: sources.PtrInt32(1),
								Items: []types.OriginCustomHeader{
									{
										HeaderName:  sources.PtrString("test-header-name"),
										HeaderValue: sources.PtrString("test-header-value"),
									},
								},
							},
							CustomOriginConfig: &types.CustomOriginConfig{
								HTTPPort:               sources.PtrInt32(80),
								HTTPSPort:              sources.PtrInt32(443),
								OriginProtocolPolicy:   types.OriginProtocolPolicyMatchViewer,
								OriginKeepaliveTimeout: sources.PtrInt32(5),
								OriginReadTimeout:      sources.PtrInt32(30),
								OriginSslProtocols: &types.OriginSslProtocols{
									Items: types.SslProtocolSSLv3.Values(),
								},
							},
							OriginAccessControlId: sources.PtrString("test-origin-access-control-id"), // link
							OriginPath:            sources.PtrString("/foo"),
							OriginShield: &types.OriginShield{
								Enabled:            sources.PtrBool(true),
								OriginShieldRegion: sources.PtrString("eu-west-1"),
							},
							S3OriginConfig: &types.S3OriginConfig{
								OriginAccessIdentity: sources.PtrString("test-origin-access-identity"), // link
							},
						},
					},
				},
				DefaultCacheBehavior: &types.DefaultCacheBehavior{
					TargetOriginId:          sources.PtrString("CustomOriginConfig"),
					ViewerProtocolPolicy:    types.ViewerProtocolPolicyHttpsOnly,
					CachePolicyId:           sources.PtrString("test-cache-policy-id"), // link
					Compress:                sources.PtrBool(true),
					DefaultTTL:              sources.PtrInt64(1),
					FieldLevelEncryptionId:  sources.PtrString("test-field-level-encryption-id"), // link
					MaxTTL:                  sources.PtrInt64(1),
					MinTTL:                  sources.PtrInt64(1),
					OriginRequestPolicyId:   sources.PtrString("test-origin-request-policy-id"),                                   // link
					RealtimeLogConfigArn:    sources.PtrString("arn:aws:logs:us-east-1:123456789012:realtime-log-config/test-id"), // link
					ResponseHeadersPolicyId: sources.PtrString("test-response-headers-policy-id"),                                 // link
					SmoothStreaming:         sources.PtrBool(true),
					ForwardedValues: &types.ForwardedValues{
						Cookies: &types.CookiePreference{
							Forward: types.ItemSelectionWhitelist,
							WhitelistedNames: &types.CookieNames{
								Quantity: sources.PtrInt32(1),
								Items: []string{
									"cooke_123",
								},
							},
						},
						QueryString: sources.PtrBool(true),
						Headers: &types.Headers{
							Quantity: sources.PtrInt32(1),
							Items: []string{
								"X-Customer-Header",
							},
						},
						QueryStringCacheKeys: &types.QueryStringCacheKeys{
							Quantity: sources.PtrInt32(1),
							Items: []string{
								"test-query-string-cache-key",
							},
						},
					},
					FunctionAssociations: &types.FunctionAssociations{
						Quantity: sources.PtrInt32(1),
						Items: []types.FunctionAssociation{
							{
								EventType:   types.EventTypeViewerRequest,
								FunctionARN: sources.PtrString("arn:aws:cloudfront::123412341234:function/1234"), // link
							},
						},
					},
					LambdaFunctionAssociations: &types.LambdaFunctionAssociations{
						Quantity: sources.PtrInt32(1),
						Items: []types.LambdaFunctionAssociation{
							{
								EventType:         types.EventTypeOriginRequest,
								LambdaFunctionARN: sources.PtrString("arn:aws:lambda:us-east-1:123456789012:function:test-function"), // link
								IncludeBody:       sources.PtrBool(true),
							},
						},
					},
					TrustedKeyGroups: &types.TrustedKeyGroups{
						Enabled:  sources.PtrBool(true),
						Quantity: sources.PtrInt32(1),
						Items: []string{
							"key-group-1", // link
						},
					},
					TrustedSigners: &types.TrustedSigners{
						Enabled:  sources.PtrBool(true),
						Quantity: sources.PtrInt32(1),
						Items: []string{
							"123456789",
						},
					},
					AllowedMethods: &types.AllowedMethods{
						Items: []types.Method{
							types.MethodGet,
						},
						Quantity: sources.PtrInt32(1),
						CachedMethods: &types.CachedMethods{
							Items: []types.Method{
								types.MethodGet,
							},
						},
					},
				},
				CustomErrorResponses: &types.CustomErrorResponses{
					Quantity: sources.PtrInt32(1),
					Items: []types.CustomErrorResponse{
						{
							ErrorCode:          sources.PtrInt32(404),
							ErrorCachingMinTTL: sources.PtrInt64(1),
							ResponseCode:       sources.PtrString("200"),
							ResponsePagePath:   sources.PtrString("/foo"),
						},
					},
				},
				DefaultRootObject: sources.PtrString("index.html"),
				HttpVersion:       types.HttpVersionHttp11,
				IsIPV6Enabled:     sources.PtrBool(true),
				Logging: &types.LoggingConfig{
					Bucket:         sources.PtrString("aws-cf-access-logs.s3.amazonaws.com"), // link
					Enabled:        sources.PtrBool(true),
					IncludeCookies: sources.PtrBool(true),
					Prefix:         sources.PtrString("test-prefix"),
				},
				OriginGroups: &types.OriginGroups{
					Quantity: sources.PtrInt32(1),
					Items: []types.OriginGroup{
						{
							FailoverCriteria: &types.OriginGroupFailoverCriteria{
								StatusCodes: &types.StatusCodes{
									Items: []int32{
										404,
									},
									Quantity: sources.PtrInt32(1),
								},
							},
							Id: sources.PtrString("test-id"),
							Members: &types.OriginGroupMembers{
								Quantity: sources.PtrInt32(1),
								Items: []types.OriginGroupMember{
									{
										OriginId: sources.PtrString("CustomOriginConfig"),
									},
								},
							},
						},
					},
				},
				PriceClass: types.PriceClassPriceClass200,
				Restrictions: &types.Restrictions{
					GeoRestriction: &types.GeoRestriction{
						Quantity:        sources.PtrInt32(1),
						RestrictionType: types.GeoRestrictionTypeWhitelist,
						Items: []string{
							"US",
						},
					},
				},
				ViewerCertificate: &types.ViewerCertificate{
					ACMCertificateArn:            sources.PtrString("arn:aws:acm:us-east-1:123456789012:certificate/test-id"), // link
					Certificate:                  sources.PtrString("test-certificate"),
					CertificateSource:            types.CertificateSourceAcm,
					CloudFrontDefaultCertificate: sources.PtrBool(true),
					IAMCertificateId:             sources.PtrString("test-iam-certificate-id"), // link
					MinimumProtocolVersion:       types.MinimumProtocolVersion(types.SslProtocolSSLv3),
					SSLSupportMethod:             types.SSLSupportMethodSniOnly,
				},
				// Note this can also be in the format: 473e64fd-f30b-4765-81a0-62ad96dd167a for WAF Classic
				WebACLId: sources.PtrString("arn:aws:wafv2:us-east-1:123456789012:global/webacl/ExampleWebACL/473e64fd-f30b-4765-81a0-62ad96dd167a"), // link
			},
		},
	}, nil
}

func (t TestCloudFrontClient) ListDistributions(ctx context.Context, params *cloudfront.ListDistributionsInput, optFns ...func(*cloudfront.Options)) (*cloudfront.ListDistributionsOutput, error) {
	return &cloudfront.ListDistributionsOutput{
		DistributionList: &types.DistributionList{
			IsTruncated: sources.PtrBool(false),
			Items: []types.DistributionSummary{
				{
					Id: sources.PtrString("test-id"),
				},
			},
		},
	}, nil
}

func TestDistributionGetFunc(t *testing.T) {
	scope := "123456789012.us-east-1"
	item, err := distributionGetFunc(context.Background(), TestCloudFrontClient{}, scope, &cloudfront.GetDistributionInput{})

	if err != nil {
		t.Fatal(err)
	}

	if err = item.Validate(); err != nil {
		t.Error(err)
	}

	if *item.Health != sdp.Health_HEALTH_OK {
		t.Errorf("expected health to be HEALTH_OK, got %s", item.Health)
	}

	tests := sources.QueryTests{
		{
			ExpectedType:   "dns",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "d111111abcdef8.cloudfront.net",
			ExpectedScope:  "global",
		},
		{
			ExpectedType:   "cloudfront-key-group",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "key-group-1",
			ExpectedScope:  scope,
		},
		{
			ExpectedType:   "dns",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "something.foo.bar.com",
			ExpectedScope:  "global",
		},
		{
			ExpectedType:   "cloudfront-continuous-deployment-policy",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "test-continuous-deployment-policy-id",
			ExpectedScope:  scope,
		},
		{
			ExpectedType:   "cloudfront-cache-policy",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "test-cache-policy-id",
			ExpectedScope:  scope,
		},
		{
			ExpectedType:   "cloudfront-field-level-encryption",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "test-field-level-encryption-id",
			ExpectedScope:  scope,
		},
		{
			ExpectedType:   "cloudfront-origin-request-policy",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "test-origin-request-policy-id",
			ExpectedScope:  scope,
		},
		{
			ExpectedType:   "cloudfront-realtime-log-config",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "arn:aws:logs:us-east-1:123456789012:realtime-log-config/test-id",
			ExpectedScope:  "123456789012.us-east-1",
		},
		{
			ExpectedType:   "cloudfront-response-headers-policy",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "test-response-headers-policy-id",
			ExpectedScope:  scope,
		},
		{
			ExpectedType:   "cloudfront-key-group",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "key-group-1",
			ExpectedScope:  scope,
		},
		{
			ExpectedType:   "cloudfront-function",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "arn:aws:cloudfront::123412341234:function/1234",
			ExpectedScope:  "123412341234",
		},
		{
			ExpectedType:   "lambda-function",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "arn:aws:lambda:us-east-1:123456789012:function:test-function",
			ExpectedScope:  "123456789012.us-east-1",
		},
		{
			ExpectedType:   "dns",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "DOC-EXAMPLE-BUCKET.s3.us-west-2.amazonaws.com",
			ExpectedScope:  "global",
		},
		{
			ExpectedType:   "cloudfront-origin-access-control",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "test-origin-access-control-id",
			ExpectedScope:  scope,
		},
		{
			ExpectedType:   "cloudfront-cloud-front-origin-access-identity",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "test-origin-access-identity",
			ExpectedScope:  scope,
		},
		{
			ExpectedType:   "dns",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "aws-cf-access-logs.s3.amazonaws.com",
			ExpectedScope:  "global",
		},
		{
			ExpectedType:   "acm-certificate",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "arn:aws:acm:us-east-1:123456789012:certificate/test-id",
			ExpectedScope:  "123456789012.us-east-1",
		},
		{
			ExpectedType:   "iam-server-certificate",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "test-iam-certificate-id",
			ExpectedScope:  scope,
		},
		{
			ExpectedType:   "wafv2-web-acl",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "arn:aws:wafv2:us-east-1:123456789012:global/webacl/ExampleWebACL/473e64fd-f30b-4765-81a0-62ad96dd167a",
			ExpectedScope:  "123456789012.us-east-1",
		},
		{
			ExpectedType:   "s3-bucket",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "DOC-EXAMPLE-BUCKET",
			ExpectedScope:  "123456789012",
		},
	}

	tests.Execute(t, item)
}

func TestNewDistributionSource(t *testing.T) {
	config, account, _ := sources.GetAutoConfig(t)

	source := NewDistributionSource(config, account)

	test := sources.E2ETest{
		Source:  source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}

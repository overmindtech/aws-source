package s3

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func TestS3SearchImpl(t *testing.T) {
	t.Run("with a good ARN", func(t *testing.T) {
		items, err := searchImpl(context.Background(), TestS3Client{}, "account-id.region", "arn:partition:service:region:account-id:resource-type:resource-id")

		if err != nil {
			t.Error(err)
		}
		if len(items) != 1 {
			t.Errorf("expected 1 item, got %v", len(items))
		}
	})

	t.Run("with a bad ARN", func(t *testing.T) {
		_, err := searchImpl(context.Background(), TestS3Client{}, "account-id.region", "foo")

		if err == nil {
			t.Error("expected error")
		} else {
			if ire, ok := err.(*sdp.QueryError); ok {
				if ire.ErrorType != sdp.QueryError_OTHER {
					t.Errorf("expected error type to be OTHER, got %v", ire.ErrorType.String())
				}
			} else {
				t.Errorf("expected item request error, got %T", err)
			}
		}
	})

	t.Run("with an ARN in another scope", func(t *testing.T) {
		_, err := searchImpl(context.Background(), TestS3Client{}, "account-id.region", "arn:partition:service:region:account-id-2:resource-type:resource-id")

		if err == nil {
			t.Error("expected error")
		} else {
			if ire, ok := err.(*sdp.QueryError); ok {
				if ire.ErrorType != sdp.QueryError_NOSCOPE {
					t.Errorf("expected error type to be OTHER, got %v", ire.ErrorType.String())
				}
			} else {
				t.Errorf("expected item request error, got %T", err)
			}
		}
	})
}

func TestS3ListImpl(t *testing.T) {
	items, err := listImpl(context.Background(), TestS3Client{}, "foo")

	if err != nil {
		t.Error(err)
	}
	if len(items) != 1 {
		t.Errorf("expected 1 item, got %v", len(items))
	}
}

func TestS3GetImpl(t *testing.T) {
	item, err := getImpl(context.Background(), TestS3Client{}, "foo", "bar")

	if err != nil {
		t.Fatal(err)
	}

	tests := sources.ItemRequestTests{
		{
			ExpectedType:   "http",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "https://hostname",
			ExpectedScope:  "global",
		},
		{
			ExpectedType:   "lambda-function",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "arn:partition:service:region:account-id:resource-type:resource-id",
			ExpectedScope:  "account-id.region",
		},
		{
			ExpectedType:   "sqs-queue",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "arn:partition:service:region:account-id:resource-type:resource-id",
			ExpectedScope:  "account-id.region",
		},
		{
			ExpectedType:   "sns-topic",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "arn:partition:service:region:account-id:resource-type:resource-id",
			ExpectedScope:  "account-id.region",
		},
		{
			ExpectedType:   "s3-bucket",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "bucket",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "ec2-region",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  string(types.BucketLocationConstraintAfSouth1),
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "s3-bucket",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "arn:partition:service:region:account-id:resource-type:resource-id",
			ExpectedScope:  "account-id.region",
		},
		{
			ExpectedType:   "s3-bucket",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "arn:partition:service:region:account-id:resource-type:resource-id",
			ExpectedScope:  "account-id.region",
		},
	}

	tests.Execute(t, item)
}

var owner = types.Owner{
	DisplayName: sources.PtrString("dylan"),
	ID:          sources.PtrString("id"),
}

// TestS3Client A client that returns example data
type TestS3Client struct{}

func (t TestS3Client) ListBuckets(ctx context.Context, params *s3.ListBucketsInput, optFns ...func(*s3.Options)) (*s3.ListBucketsOutput, error) {
	return &s3.ListBucketsOutput{
		Buckets: []types.Bucket{
			{
				CreationDate: sources.PtrTime(time.Now()),
				Name:         sources.PtrString("foo"),
			},
		},
		Owner: &owner,
	}, nil
}

func (t TestS3Client) GetBucketAcl(ctx context.Context, params *s3.GetBucketAclInput, optFns ...func(*s3.Options)) (*s3.GetBucketAclOutput, error) {
	return &s3.GetBucketAclOutput{
		Grants: []types.Grant{
			{
				Grantee: &types.Grantee{
					Type:         types.TypeAmazonCustomerByEmail,
					DisplayName:  sources.PtrString("dylan"),
					EmailAddress: sources.PtrString("dylan@company.com"),
					ID:           sources.PtrString("id"),
					URI:          sources.PtrString("uri"),
				},
			},
		},
		Owner: &owner,
	}, nil
}

func (t TestS3Client) GetBucketAnalyticsConfiguration(ctx context.Context, params *s3.GetBucketAnalyticsConfigurationInput, optFns ...func(*s3.Options)) (*s3.GetBucketAnalyticsConfigurationOutput, error) {
	return &s3.GetBucketAnalyticsConfigurationOutput{
		AnalyticsConfiguration: &types.AnalyticsConfiguration{
			Id: sources.PtrString("id"),
			StorageClassAnalysis: &types.StorageClassAnalysis{
				DataExport: &types.StorageClassAnalysisDataExport{
					Destination: &types.AnalyticsExportDestination{
						S3BucketDestination: &types.AnalyticsS3BucketDestination{
							Bucket:          sources.PtrString("arn:partition:service:region:account-id:resource-type:resource-id"),
							Format:          types.AnalyticsS3ExportFileFormatCsv,
							BucketAccountId: sources.PtrString("id"),
							Prefix:          sources.PtrString("pre"),
						},
					},
					OutputSchemaVersion: types.StorageClassAnalysisSchemaVersionV1,
				},
			},
		},
	}, nil
}

func (t TestS3Client) GetBucketCors(ctx context.Context, params *s3.GetBucketCorsInput, optFns ...func(*s3.Options)) (*s3.GetBucketCorsOutput, error) {
	return &s3.GetBucketCorsOutput{
		CORSRules: []types.CORSRule{
			{
				AllowedMethods: []string{
					"GET",
				},
				AllowedOrigins: []string{
					"amazon.com",
				},
				AllowedHeaders: []string{
					"Authorization",
				},
				ExposeHeaders: []string{
					"foo",
				},
				ID:            sources.PtrString("id"),
				MaxAgeSeconds: 10,
			},
		},
	}, nil
}

func (t TestS3Client) GetBucketEncryption(ctx context.Context, params *s3.GetBucketEncryptionInput, optFns ...func(*s3.Options)) (*s3.GetBucketEncryptionOutput, error) {
	return &s3.GetBucketEncryptionOutput{
		ServerSideEncryptionConfiguration: &types.ServerSideEncryptionConfiguration{
			Rules: []types.ServerSideEncryptionRule{
				{
					ApplyServerSideEncryptionByDefault: &types.ServerSideEncryptionByDefault{
						SSEAlgorithm:   types.ServerSideEncryptionAes256,
						KMSMasterKeyID: sources.PtrString("id"),
					},
					BucketKeyEnabled: true,
				},
			},
		},
	}, nil
}

func (t TestS3Client) GetBucketIntelligentTieringConfiguration(ctx context.Context, params *s3.GetBucketIntelligentTieringConfigurationInput, optFns ...func(*s3.Options)) (*s3.GetBucketIntelligentTieringConfigurationOutput, error) {
	return &s3.GetBucketIntelligentTieringConfigurationOutput{
		IntelligentTieringConfiguration: &types.IntelligentTieringConfiguration{
			Id:     sources.PtrString("id"),
			Status: types.IntelligentTieringStatusEnabled,
			Tierings: []types.Tiering{
				{
					AccessTier: types.IntelligentTieringAccessTierDeepArchiveAccess,
					Days:       100,
				},
			},
			Filter: &types.IntelligentTieringFilter{},
		},
	}, nil
}

func (t TestS3Client) GetBucketInventoryConfiguration(ctx context.Context, params *s3.GetBucketInventoryConfigurationInput, optFns ...func(*s3.Options)) (*s3.GetBucketInventoryConfigurationOutput, error) {
	return &s3.GetBucketInventoryConfigurationOutput{
		InventoryConfiguration: &types.InventoryConfiguration{
			Destination: &types.InventoryDestination{
				S3BucketDestination: &types.InventoryS3BucketDestination{
					Bucket:    sources.PtrString("arn:partition:service:region:account-id:resource-type:resource-id"),
					Format:    types.InventoryFormatCsv,
					AccountId: sources.PtrString("id"),
					Encryption: &types.InventoryEncryption{
						SSEKMS: &types.SSEKMS{
							KeyId: sources.PtrString("key"),
						},
					},
					Prefix: sources.PtrString("pre"),
				},
			},
			Id:                     sources.PtrString("id"),
			IncludedObjectVersions: types.InventoryIncludedObjectVersionsAll,
			IsEnabled:              true,
			Schedule: &types.InventorySchedule{
				Frequency: types.InventoryFrequencyDaily,
			},
		},
	}, nil
}

func (t TestS3Client) GetBucketLifecycleConfiguration(ctx context.Context, params *s3.GetBucketLifecycleConfigurationInput, optFns ...func(*s3.Options)) (*s3.GetBucketLifecycleConfigurationOutput, error) {
	return &s3.GetBucketLifecycleConfigurationOutput{
		Rules: []types.LifecycleRule{
			{
				Status: types.ExpirationStatusEnabled,
				AbortIncompleteMultipartUpload: &types.AbortIncompleteMultipartUpload{
					DaysAfterInitiation: 1,
				},
				Expiration: &types.LifecycleExpiration{
					Date:                      sources.PtrTime(time.Now()),
					Days:                      3,
					ExpiredObjectDeleteMarker: true,
				},
				ID: sources.PtrString("id"),
				NoncurrentVersionExpiration: &types.NoncurrentVersionExpiration{
					NewerNoncurrentVersions: 3,
					NoncurrentDays:          1,
				},
				NoncurrentVersionTransitions: []types.NoncurrentVersionTransition{
					{
						NewerNoncurrentVersions: 1,
						NoncurrentDays:          1,
						StorageClass:            types.TransitionStorageClassGlacierIr,
					},
				},
				Prefix: sources.PtrString("pre"),
				Transitions: []types.Transition{
					{
						Date:         sources.PtrTime(time.Now()),
						Days:         12,
						StorageClass: types.TransitionStorageClassGlacierIr,
					},
				},
			},
		},
	}, nil
}

func (t TestS3Client) GetBucketLocation(ctx context.Context, params *s3.GetBucketLocationInput, optFns ...func(*s3.Options)) (*s3.GetBucketLocationOutput, error) {
	return &s3.GetBucketLocationOutput{
		LocationConstraint: types.BucketLocationConstraintAfSouth1,
	}, nil
}

func (t TestS3Client) GetBucketLogging(ctx context.Context, params *s3.GetBucketLoggingInput, optFns ...func(*s3.Options)) (*s3.GetBucketLoggingOutput, error) {
	return &s3.GetBucketLoggingOutput{
		LoggingEnabled: &types.LoggingEnabled{
			TargetBucket: sources.PtrString("bucket"),
			TargetPrefix: sources.PtrString("pre"),
			TargetGrants: []types.TargetGrant{
				{
					Grantee: &types.Grantee{
						Type: types.TypeGroup,
						ID:   sources.PtrString("id"),
					},
				},
			},
		},
	}, nil
}

func (t TestS3Client) GetBucketMetricsConfiguration(ctx context.Context, params *s3.GetBucketMetricsConfigurationInput, optFns ...func(*s3.Options)) (*s3.GetBucketMetricsConfigurationOutput, error) {
	return &s3.GetBucketMetricsConfigurationOutput{
		MetricsConfiguration: &types.MetricsConfiguration{
			Id: sources.PtrString("id"),
		},
	}, nil
}

func (t TestS3Client) GetBucketNotificationConfiguration(ctx context.Context, params *s3.GetBucketNotificationConfigurationInput, optFns ...func(*s3.Options)) (*s3.GetBucketNotificationConfigurationOutput, error) {
	return &s3.GetBucketNotificationConfigurationOutput{
		LambdaFunctionConfigurations: []types.LambdaFunctionConfiguration{
			{
				Events:            []types.Event{},
				LambdaFunctionArn: sources.PtrString("arn:partition:service:region:account-id:resource-type:resource-id"),
				Id:                sources.PtrString("id"),
			},
		},
		EventBridgeConfiguration: &types.EventBridgeConfiguration{},
		QueueConfigurations: []types.QueueConfiguration{
			{
				Events:   []types.Event{},
				QueueArn: sources.PtrString("arn:partition:service:region:account-id:resource-type:resource-id"),
				Filter: &types.NotificationConfigurationFilter{
					Key: &types.S3KeyFilter{
						FilterRules: []types.FilterRule{
							{
								Name:  types.FilterRuleNamePrefix,
								Value: sources.PtrString("foo"),
							},
						},
					},
				},
				Id: sources.PtrString("id"),
			},
		},
		TopicConfigurations: []types.TopicConfiguration{
			{
				Events:   []types.Event{},
				TopicArn: sources.PtrString("arn:partition:service:region:account-id:resource-type:resource-id"),
				Filter: &types.NotificationConfigurationFilter{
					Key: &types.S3KeyFilter{
						FilterRules: []types.FilterRule{
							{
								Name:  types.FilterRuleNameSuffix,
								Value: sources.PtrString("fix"),
							},
						},
					},
				},
				Id: sources.PtrString("id"),
			},
		},
	}, nil
}

func (t TestS3Client) GetBucketOwnershipControls(ctx context.Context, params *s3.GetBucketOwnershipControlsInput, optFns ...func(*s3.Options)) (*s3.GetBucketOwnershipControlsOutput, error) {
	return &s3.GetBucketOwnershipControlsOutput{
		OwnershipControls: &types.OwnershipControls{
			Rules: []types.OwnershipControlsRule{
				{
					ObjectOwnership: types.ObjectOwnershipBucketOwnerPreferred,
				},
			},
		},
	}, nil
}

func (t TestS3Client) GetBucketPolicy(ctx context.Context, params *s3.GetBucketPolicyInput, optFns ...func(*s3.Options)) (*s3.GetBucketPolicyOutput, error) {
	return &s3.GetBucketPolicyOutput{
		Policy: sources.PtrString("policy"),
	}, nil
}

func (t TestS3Client) GetBucketPolicyStatus(ctx context.Context, params *s3.GetBucketPolicyStatusInput, optFns ...func(*s3.Options)) (*s3.GetBucketPolicyStatusOutput, error) {
	return &s3.GetBucketPolicyStatusOutput{
		PolicyStatus: &types.PolicyStatus{
			IsPublic: true,
		},
	}, nil
}

func (t TestS3Client) GetBucketReplication(ctx context.Context, params *s3.GetBucketReplicationInput, optFns ...func(*s3.Options)) (*s3.GetBucketReplicationOutput, error) {
	return &s3.GetBucketReplicationOutput{
		ReplicationConfiguration: &types.ReplicationConfiguration{
			Role: sources.PtrString("role"),
			Rules: []types.ReplicationRule{
				{
					Destination: &types.Destination{
						Bucket: sources.PtrString("bucket"),
						AccessControlTranslation: &types.AccessControlTranslation{
							Owner: types.OwnerOverrideDestination,
						},
						Account: sources.PtrString("account"),
						EncryptionConfiguration: &types.EncryptionConfiguration{
							ReplicaKmsKeyID: sources.PtrString("keyId"),
						},
						Metrics: &types.Metrics{
							Status: types.MetricsStatusEnabled,
							EventThreshold: &types.ReplicationTimeValue{
								Minutes: 1,
							},
						},
						ReplicationTime: &types.ReplicationTime{
							Status: types.ReplicationTimeStatusEnabled,
							Time: &types.ReplicationTimeValue{
								Minutes: 1,
							},
						},
						StorageClass: types.StorageClassGlacier,
					},
				},
			},
		},
	}, nil
}

func (t TestS3Client) GetBucketRequestPayment(ctx context.Context, params *s3.GetBucketRequestPaymentInput, optFns ...func(*s3.Options)) (*s3.GetBucketRequestPaymentOutput, error) {
	return &s3.GetBucketRequestPaymentOutput{
		Payer: types.PayerRequester,
	}, nil
}

func (t TestS3Client) GetBucketTagging(ctx context.Context, params *s3.GetBucketTaggingInput, optFns ...func(*s3.Options)) (*s3.GetBucketTaggingOutput, error) {
	return &s3.GetBucketTaggingOutput{
		TagSet: []types.Tag{},
	}, nil
}

func (t TestS3Client) GetBucketVersioning(ctx context.Context, params *s3.GetBucketVersioningInput, optFns ...func(*s3.Options)) (*s3.GetBucketVersioningOutput, error) {
	return &s3.GetBucketVersioningOutput{
		MFADelete: types.MFADeleteStatusEnabled,
		Status:    types.BucketVersioningStatusSuspended,
	}, nil
}

func (t TestS3Client) GetBucketWebsite(ctx context.Context, params *s3.GetBucketWebsiteInput, optFns ...func(*s3.Options)) (*s3.GetBucketWebsiteOutput, error) {
	return &s3.GetBucketWebsiteOutput{
		ErrorDocument: &types.ErrorDocument{
			Key: sources.PtrString("key"),
		},
		IndexDocument: &types.IndexDocument{
			Suffix: sources.PtrString("html"),
		},
		RedirectAllRequestsTo: &types.RedirectAllRequestsTo{
			HostName: sources.PtrString("hostname"),
			Protocol: types.ProtocolHttps,
		},
		RoutingRules: []types.RoutingRule{
			{
				Redirect: &types.Redirect{
					HostName:             sources.PtrString("hostname"),
					HttpRedirectCode:     sources.PtrString("303"),
					Protocol:             types.ProtocolHttp,
					ReplaceKeyPrefixWith: sources.PtrString("pre"),
					ReplaceKeyWith:       sources.PtrString("key"),
				},
			},
		},
	}, nil
}

func TestNewS3Source(t *testing.T) {
	config, account, _ := sources.GetAutoConfig(t)

	source := NewS3Source(config, account)

	test := sources.E2ETest{
		Source:  source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}

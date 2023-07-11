package s3

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/getsentry/sentry-go"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

// NewS3Source Creates a new S3 source
func NewS3Source(config aws.Config, accountID string) *S3Source {
	return &S3Source{
		config:    config,
		accountID: accountID,
	}
}

//go:generate docgen ../../docs-data
// +overmind:descriptiveType S3 Bucket
// +overmind:get Get an S3 bucket by name
// +overmind:list List all S3 buckets
// +overmind:search Search for S3 buckets by ARN
// +overmind:group AWS
// +overmind:terraform:queryMap aws_s3_bucket_acl.bucket
// +overmind:terraform:queryMap aws_s3_bucket_analytics_configuration.bucket
// +overmind:terraform:queryMap aws_s3_bucket_cors_configuration.bucket
// +overmind:terraform:queryMap aws_s3_bucket_intelligent_tiering_configuration.bucket
// +overmind:terraform:queryMap aws_s3_bucket_inventory.bucket
// +overmind:terraform:queryMap aws_s3_bucket_lifecycle_configuration.bucket
// +overmind:terraform:queryMap aws_s3_bucket_logging.bucket
// +overmind:terraform:queryMap aws_s3_bucket_metric.bucket
// +overmind:terraform:queryMap aws_s3_bucket_notification.bucket
// +overmind:terraform:queryMap aws_s3_bucket_object_lock_configuration.bucket
// +overmind:terraform:queryMap aws_s3_bucket_object.bucket
// +overmind:terraform:queryMap aws_s3_bucket_ownership_controls.bucket
// +overmind:terraform:queryMap aws_s3_bucket_policy.bucket
// +overmind:terraform:queryMap aws_s3_bucket_public_access_block.bucket
// +overmind:terraform:queryMap aws_s3_bucket_replication_configuration.bucket
// +overmind:terraform:queryMap aws_s3_bucket_request_payment_configuration.bucket
// +overmind:terraform:queryMap aws_s3_bucket_server_side_encryption_configuration.bucket
// +overmind:terraform:queryMap aws_s3_bucket_versioning.bucket
// +overmind:terraform:queryMap aws_s3_bucket_website_configuration.bucket
// +overmind:terraform:queryMap aws_s3_bucket.id
// +overmind:terraform:queryMap aws_s3_object_copy.bucket
// +overmind:terraform:queryMap aws_s3_object.bucket

type S3Source struct {
	// AWS Config including region and credentials
	config aws.Config

	// AccountID The id of the account that is being used. This is used by
	// sources as the first element in the scope
	accountID string

	// client The AWS client to use when making requests
	client        *s3.Client
	clientCreated bool
	clientMutex   sync.Mutex
}

func (s *S3Source) Client() *s3.Client {
	s.clientMutex.Lock()
	defer s.clientMutex.Unlock()

	// If the client already exists then return it
	if s.clientCreated {
		return s.client
	}

	// Otherwise create a new client from the config
	s.client = s3.NewFromConfig(s.config)
	s.clientCreated = true

	return s.client
}

// Type The type of items that this source is capable of finding
func (s *S3Source) Type() string {
	// +overmind:type s3-bucket
	return "s3-bucket"
}

// Descriptive name for the source, used in logging and metadata
func (s *S3Source) Name() string {
	return "aws-s3-source"
}

// List of scopes that this source is capable of find items for. This will be
// in the format {accountID}.{region}
func (s *S3Source) Scopes() []string {
	return []string{
		sources.FormatScope(s.accountID, s.config.Region),
	}
}

// S3Client A client that can get data about S3 buckets
type S3Client interface {
	ListBuckets(ctx context.Context, params *s3.ListBucketsInput, optFns ...func(*s3.Options)) (*s3.ListBucketsOutput, error)
	GetBucketAcl(ctx context.Context, params *s3.GetBucketAclInput, optFns ...func(*s3.Options)) (*s3.GetBucketAclOutput, error)
	GetBucketAnalyticsConfiguration(ctx context.Context, params *s3.GetBucketAnalyticsConfigurationInput, optFns ...func(*s3.Options)) (*s3.GetBucketAnalyticsConfigurationOutput, error)
	GetBucketCors(ctx context.Context, params *s3.GetBucketCorsInput, optFns ...func(*s3.Options)) (*s3.GetBucketCorsOutput, error)
	GetBucketEncryption(ctx context.Context, params *s3.GetBucketEncryptionInput, optFns ...func(*s3.Options)) (*s3.GetBucketEncryptionOutput, error)
	GetBucketIntelligentTieringConfiguration(ctx context.Context, params *s3.GetBucketIntelligentTieringConfigurationInput, optFns ...func(*s3.Options)) (*s3.GetBucketIntelligentTieringConfigurationOutput, error)
	GetBucketInventoryConfiguration(ctx context.Context, params *s3.GetBucketInventoryConfigurationInput, optFns ...func(*s3.Options)) (*s3.GetBucketInventoryConfigurationOutput, error)
	GetBucketLifecycleConfiguration(ctx context.Context, params *s3.GetBucketLifecycleConfigurationInput, optFns ...func(*s3.Options)) (*s3.GetBucketLifecycleConfigurationOutput, error)
	GetBucketLocation(ctx context.Context, params *s3.GetBucketLocationInput, optFns ...func(*s3.Options)) (*s3.GetBucketLocationOutput, error)
	GetBucketLogging(ctx context.Context, params *s3.GetBucketLoggingInput, optFns ...func(*s3.Options)) (*s3.GetBucketLoggingOutput, error)
	GetBucketMetricsConfiguration(ctx context.Context, params *s3.GetBucketMetricsConfigurationInput, optFns ...func(*s3.Options)) (*s3.GetBucketMetricsConfigurationOutput, error)
	GetBucketNotificationConfiguration(ctx context.Context, params *s3.GetBucketNotificationConfigurationInput, optFns ...func(*s3.Options)) (*s3.GetBucketNotificationConfigurationOutput, error)
	GetBucketOwnershipControls(ctx context.Context, params *s3.GetBucketOwnershipControlsInput, optFns ...func(*s3.Options)) (*s3.GetBucketOwnershipControlsOutput, error)
	GetBucketPolicy(ctx context.Context, params *s3.GetBucketPolicyInput, optFns ...func(*s3.Options)) (*s3.GetBucketPolicyOutput, error)
	GetBucketPolicyStatus(ctx context.Context, params *s3.GetBucketPolicyStatusInput, optFns ...func(*s3.Options)) (*s3.GetBucketPolicyStatusOutput, error)
	GetBucketReplication(ctx context.Context, params *s3.GetBucketReplicationInput, optFns ...func(*s3.Options)) (*s3.GetBucketReplicationOutput, error)
	GetBucketRequestPayment(ctx context.Context, params *s3.GetBucketRequestPaymentInput, optFns ...func(*s3.Options)) (*s3.GetBucketRequestPaymentOutput, error)
	GetBucketTagging(ctx context.Context, params *s3.GetBucketTaggingInput, optFns ...func(*s3.Options)) (*s3.GetBucketTaggingOutput, error)
	GetBucketVersioning(ctx context.Context, params *s3.GetBucketVersioningInput, optFns ...func(*s3.Options)) (*s3.GetBucketVersioningOutput, error)
	GetBucketWebsite(ctx context.Context, params *s3.GetBucketWebsiteInput, optFns ...func(*s3.Options)) (*s3.GetBucketWebsiteOutput, error)
}

// Bucket represents an actual s3 bucket, with all of the extra requests
// resolved and all information added
type Bucket struct {
	// ListBuckets
	types.Bucket

	s3.GetBucketAclOutput
	s3.GetBucketAnalyticsConfigurationOutput
	s3.GetBucketCorsOutput
	s3.GetBucketEncryptionOutput
	s3.GetBucketIntelligentTieringConfigurationOutput
	s3.GetBucketInventoryConfigurationOutput
	s3.GetBucketLifecycleConfigurationOutput
	s3.GetBucketLocationOutput
	s3.GetBucketLoggingOutput
	s3.GetBucketMetricsConfigurationOutput
	s3.GetBucketNotificationConfigurationOutput
	s3.GetBucketOwnershipControlsOutput
	s3.GetBucketPolicyOutput
	s3.GetBucketPolicyStatusOutput
	s3.GetBucketReplicationOutput
	s3.GetBucketRequestPaymentOutput
	s3.GetBucketTaggingOutput
	s3.GetBucketVersioningOutput
	s3.GetBucketWebsiteOutput
}

// Get Get a single item with a given scope and query. The item returned
// should have a UniqueAttributeValue that matches the `query` parameter. The
// ctx parameter contains a golang context object which should be used to allow
// this source to timeout or be cancelled when executing potentially
// long-running actions
func (s *S3Source) Get(ctx context.Context, scope string, query string) (*sdp.Item, error) {
	if scope != s.Scopes()[0] {
		return nil, &sdp.QueryError{
			ErrorType:   sdp.QueryError_NOSCOPE,
			ErrorString: fmt.Sprintf("requested scope %v does not match source scope %v", scope, s.Scopes()[0]),
			Scope:       scope,
		}
	}

	return getImpl(ctx, s.Client(), scope, query)
}

func getImpl(ctx context.Context, client S3Client, scope string, query string) (*sdp.Item, error) {
	var location *s3.GetBucketLocationOutput
	var wg sync.WaitGroup
	var err error

	bucketName := sources.PtrString(query)

	location, err = client.GetBucketLocation(ctx, &s3.GetBucketLocationInput{
		Bucket: bucketName,
	})

	if err != nil {
		return nil, sources.WrapAWSError(err)
	}

	bucket := Bucket{
		Bucket: types.Bucket{
			Name: bucketName,
		},
		GetBucketLocationOutput: *location,
	}

	// We want to execute all of these requests in parallel so we're not
	// crippled by latency. This API is really stupid but there's not much I can
	// do about it

	wg.Add(1)
	go func() {
		defer sentry.Recover()
		defer wg.Done()
		if acl, err := client.GetBucketAcl(ctx, &s3.GetBucketAclInput{Bucket: bucketName}); err == nil {
			bucket.GetBucketAclOutput = *acl
		}
	}()
	wg.Add(1)
	go func() {
		defer sentry.Recover()
		defer wg.Done()
		if analyticsConfiguration, err := client.GetBucketAnalyticsConfiguration(ctx, &s3.GetBucketAnalyticsConfigurationInput{Bucket: bucketName}); err == nil {
			bucket.GetBucketAnalyticsConfigurationOutput = *analyticsConfiguration
		}
	}()
	wg.Add(1)
	go func() {
		defer sentry.Recover()
		defer wg.Done()
		if cors, err := client.GetBucketCors(ctx, &s3.GetBucketCorsInput{Bucket: bucketName}); err == nil {
			bucket.GetBucketCorsOutput = *cors
		}
	}()
	wg.Add(1)
	go func() {
		defer sentry.Recover()
		defer wg.Done()
		if encryption, err := client.GetBucketEncryption(ctx, &s3.GetBucketEncryptionInput{Bucket: bucketName}); err == nil {
			bucket.GetBucketEncryptionOutput = *encryption
		}
	}()
	wg.Add(1)
	go func() {
		defer sentry.Recover()
		defer wg.Done()
		if intelligentTieringConfiguration, err := client.GetBucketIntelligentTieringConfiguration(ctx, &s3.GetBucketIntelligentTieringConfigurationInput{Bucket: bucketName}); err == nil {
			bucket.GetBucketIntelligentTieringConfigurationOutput = *intelligentTieringConfiguration
		}
	}()
	wg.Add(1)
	go func() {
		defer sentry.Recover()
		defer wg.Done()
		if inventoryConfiguration, err := client.GetBucketInventoryConfiguration(ctx, &s3.GetBucketInventoryConfigurationInput{Bucket: bucketName}); err == nil {
			bucket.GetBucketInventoryConfigurationOutput = *inventoryConfiguration
		}
	}()
	wg.Add(1)
	go func() {
		defer sentry.Recover()
		defer wg.Done()
		if lifecycleConfiguration, err := client.GetBucketLifecycleConfiguration(ctx, &s3.GetBucketLifecycleConfigurationInput{Bucket: bucketName}); err == nil {
			bucket.GetBucketLifecycleConfigurationOutput = *lifecycleConfiguration
		}
	}()
	wg.Add(1)
	go func() {
		defer sentry.Recover()
		defer wg.Done()
		if logging, err := client.GetBucketLogging(ctx, &s3.GetBucketLoggingInput{Bucket: bucketName}); err == nil {
			bucket.GetBucketLoggingOutput = *logging
		}
	}()
	wg.Add(1)
	go func() {
		defer sentry.Recover()
		defer wg.Done()
		if metricsConfiguration, err := client.GetBucketMetricsConfiguration(ctx, &s3.GetBucketMetricsConfigurationInput{Bucket: bucketName}); err == nil {
			bucket.GetBucketMetricsConfigurationOutput = *metricsConfiguration
		}
	}()
	wg.Add(1)
	go func() {
		defer sentry.Recover()
		defer wg.Done()
		if notificationConfiguration, err := client.GetBucketNotificationConfiguration(ctx, &s3.GetBucketNotificationConfigurationInput{Bucket: bucketName}); err == nil {
			bucket.GetBucketNotificationConfigurationOutput = *notificationConfiguration
		}
	}()
	wg.Add(1)
	go func() {
		defer sentry.Recover()
		defer wg.Done()
		if ownershipControls, err := client.GetBucketOwnershipControls(ctx, &s3.GetBucketOwnershipControlsInput{Bucket: bucketName}); err == nil {
			bucket.GetBucketOwnershipControlsOutput = *ownershipControls
		}
	}()
	wg.Add(1)
	go func() {
		defer sentry.Recover()
		defer wg.Done()
		if policy, err := client.GetBucketPolicy(ctx, &s3.GetBucketPolicyInput{Bucket: bucketName}); err == nil {
			bucket.GetBucketPolicyOutput = *policy
		}
	}()
	wg.Add(1)
	go func() {
		defer sentry.Recover()
		defer wg.Done()
		if policyStatus, err := client.GetBucketPolicyStatus(ctx, &s3.GetBucketPolicyStatusInput{Bucket: bucketName}); err == nil {
			bucket.GetBucketPolicyStatusOutput = *policyStatus
		}
	}()
	wg.Add(1)
	go func() {
		defer sentry.Recover()
		defer wg.Done()
		if replication, err := client.GetBucketReplication(ctx, &s3.GetBucketReplicationInput{Bucket: bucketName}); err == nil {
			bucket.GetBucketReplicationOutput = *replication
		}
	}()
	wg.Add(1)
	go func() {
		defer sentry.Recover()
		defer wg.Done()
		if requestPayment, err := client.GetBucketRequestPayment(ctx, &s3.GetBucketRequestPaymentInput{Bucket: bucketName}); err == nil {
			bucket.GetBucketRequestPaymentOutput = *requestPayment
		}
	}()
	wg.Add(1)
	go func() {
		defer sentry.Recover()
		defer wg.Done()
		if tagging, err := client.GetBucketTagging(ctx, &s3.GetBucketTaggingInput{Bucket: bucketName}); err == nil {
			bucket.GetBucketTaggingOutput = *tagging
		}
	}()
	wg.Add(1)
	go func() {
		defer sentry.Recover()
		defer wg.Done()
		if versioning, err := client.GetBucketVersioning(ctx, &s3.GetBucketVersioningInput{Bucket: bucketName}); err == nil {
			bucket.GetBucketVersioningOutput = *versioning
		}
	}()
	wg.Add(1)
	go func() {
		defer sentry.Recover()
		defer wg.Done()
		if website, err := client.GetBucketWebsite(ctx, &s3.GetBucketWebsiteInput{Bucket: bucketName}); err == nil {
			bucket.GetBucketWebsiteOutput = *website
		}
	}()

	// Wait for all requests to complete
	wg.Wait()

	attributes, err := sources.ToAttributesCase(bucket)

	if err != nil {
		return nil, &sdp.QueryError{
			ErrorType:   sdp.QueryError_OTHER,
			ErrorString: err.Error(),
			Scope:       scope,
		}
	}

	item := sdp.Item{
		Type:            "s3-bucket",
		UniqueAttribute: "name",
		Attributes:      attributes,
		Scope:           scope,
	}

	if bucket.RedirectAllRequestsTo != nil {
		if bucket.RedirectAllRequestsTo.HostName != nil {
			var url string

			switch bucket.RedirectAllRequestsTo.Protocol {
			case types.ProtocolHttp:
				url = "https://" + *bucket.RedirectAllRequestsTo.HostName
			case types.ProtocolHttps:
				url = "https://" + *bucket.RedirectAllRequestsTo.HostName
			}

			// +overmind:link http
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "http",
					Method: sdp.QueryMethod_GET,
					Query:  url,
					Scope:  "global",
				},
				BlastPropagation: &sdp.BlastPropagation{
					// HTTP always linked
					In:  true,
					Out: true,
				},
			})
		}
	}

	var a *sources.ARN

	for _, lambdaConfig := range bucket.LambdaFunctionConfigurations {
		if lambdaConfig.LambdaFunctionArn != nil {
			if a, err = sources.ParseARN(*lambdaConfig.LambdaFunctionArn); err == nil {
				// +overmind:link lambda-function
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "lambda-function",
						Method: sdp.QueryMethod_SEARCH,
						Query:  *lambdaConfig.LambdaFunctionArn,
						Scope:  sources.FormatScope(a.AccountID, a.Region),
					},
					BlastPropagation: &sdp.BlastPropagation{
						// Tightly coupled
						In:  true,
						Out: true,
					},
				})
			}
		}
	}

	for _, q := range bucket.QueueConfigurations {
		if q.QueueArn != nil {
			if a, err = sources.ParseARN(*q.QueueArn); err == nil {
				// +overmind:link sqs-queue
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "sqs-queue",
						Method: sdp.QueryMethod_SEARCH,
						Query:  *q.QueueArn,
						Scope:  sources.FormatScope(a.AccountID, a.Region),
					},
					BlastPropagation: &sdp.BlastPropagation{
						// Tightly coupled
						In:  true,
						Out: true,
					},
				})
			}
		}
	}

	for _, topic := range bucket.TopicConfigurations {
		if topic.TopicArn != nil {
			if a, err = sources.ParseARN(*topic.TopicArn); err == nil {
				// +overmind:link sns-topic
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "sns-topic",
						Method: sdp.QueryMethod_SEARCH,
						Query:  *topic.TopicArn,
						Scope:  sources.FormatScope(a.AccountID, a.Region),
					},
					BlastPropagation: &sdp.BlastPropagation{
						// Tightly coupled
						In:  true,
						Out: true,
					},
				})
			}
		}
	}

	if bucket.LoggingEnabled != nil {
		if bucket.LoggingEnabled.TargetBucket != nil {
			// +overmind:link s3-bucket
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "s3-bucket",
					Method: sdp.QueryMethod_GET,
					Query:  *bucket.LoggingEnabled.TargetBucket,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					// Tightly coupled
					In:  true,
					Out: true,
				},
			})
		}
	}

	if bucket.LocationConstraint != "" {
		// +overmind:link ec2-region
		item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
			Query: &sdp.Query{
				Type:   "ec2-region",
				Method: sdp.QueryMethod_GET,
				Query:  string(bucket.LocationConstraint),
				Scope:  scope,
			},
			BlastPropagation: &sdp.BlastPropagation{
				// Changing the region will affect the bucket
				In: true,
				// Changing the bucket won't affect the region
				Out: false,
			},
		})
	}

	if bucket.InventoryConfiguration != nil {
		if bucket.InventoryConfiguration.Destination != nil {
			if bucket.InventoryConfiguration.Destination.S3BucketDestination != nil {
				if bucket.InventoryConfiguration.Destination.S3BucketDestination.Bucket != nil {
					if a, err = sources.ParseARN(*bucket.InventoryConfiguration.Destination.S3BucketDestination.Bucket); err == nil {
						// +overmind:link s3-bucket
						item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
							Query: &sdp.Query{
								Type:   "s3-bucket",
								Method: sdp.QueryMethod_SEARCH,
								Query:  *bucket.InventoryConfiguration.Destination.S3BucketDestination.Bucket,
								Scope:  sources.FormatScope(a.AccountID, a.Region),
							},
							BlastPropagation: &sdp.BlastPropagation{
								// Tightly coupled
								In:  true,
								Out: true,
							},
						})
					}
				}
			}
		}
	}

	// Dear god there has to be a better way to do this? Should we just let it
	// panic and then deal with it?
	if bucket.AnalyticsConfiguration != nil {
		if bucket.AnalyticsConfiguration.StorageClassAnalysis != nil {
			if bucket.AnalyticsConfiguration.StorageClassAnalysis.DataExport != nil {
				if bucket.AnalyticsConfiguration.StorageClassAnalysis.DataExport.Destination != nil {
					if bucket.AnalyticsConfiguration.StorageClassAnalysis.DataExport.Destination.S3BucketDestination != nil {
						if bucket.AnalyticsConfiguration.StorageClassAnalysis.DataExport.Destination.S3BucketDestination.Bucket != nil {
							if a, err = sources.ParseARN(*bucket.AnalyticsConfiguration.StorageClassAnalysis.DataExport.Destination.S3BucketDestination.Bucket); err == nil {
								// +overmind:link s3-bucket
								item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
									Query: &sdp.Query{
										Type:   "s3-bucket",
										Method: sdp.QueryMethod_SEARCH,
										Query:  *bucket.AnalyticsConfiguration.StorageClassAnalysis.DataExport.Destination.S3BucketDestination.Bucket,
										Scope:  sources.FormatScope(a.AccountID, a.Region),
									},
									BlastPropagation: &sdp.BlastPropagation{
										// Tightly coupled
										In:  true,
										Out: true,
									},
								})
							}
						}
					}
				}
			}
		}
	}

	return &item, nil
}

// List Lists all items in a given scope
func (s *S3Source) List(ctx context.Context, scope string) ([]*sdp.Item, error) {
	if scope != s.Scopes()[0] {
		return nil, &sdp.QueryError{
			ErrorType:   sdp.QueryError_NOSCOPE,
			ErrorString: fmt.Sprintf("requested scope %v does not match source scope %v", scope, s.Scopes()[0]),
			Scope:       scope,
		}
	}

	return listImpl(ctx, s.Client(), scope)
}

func listImpl(ctx context.Context, client S3Client, scope string) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	buckets, err := client.ListBuckets(ctx, &s3.ListBucketsInput{})

	if err != nil {
		return nil, sdp.NewQueryError(err)
	}

	for _, bucket := range buckets.Buckets {
		item, err := getImpl(ctx, client, scope, *bucket.Name)

		if err != nil {
			continue
		}

		items = append(items, item)
	}

	return items, nil
}

// Search Searches for an S3 bucket by ARN rather than name
func (s *S3Source) Search(ctx context.Context, scope string, query string) ([]*sdp.Item, error) {
	if scope != s.Scopes()[0] {
		return nil, &sdp.QueryError{
			ErrorType:   sdp.QueryError_NOSCOPE,
			ErrorString: fmt.Sprintf("requested scope %v does not match source scope %v", scope, s.Scopes()[0]),
			Scope:       scope,
		}
	}

	return searchImpl(ctx, s.client, scope, query)
}

func searchImpl(ctx context.Context, client S3Client, scope string, query string) ([]*sdp.Item, error) {
	// Parse the ARN
	a, err := sources.ParseARN(query)

	if err != nil {
		return nil, sdp.NewQueryError(err)
	}

	if arnScope := sources.FormatScope(a.AccountID, a.Region); arnScope != scope {
		return nil, &sdp.QueryError{
			ErrorType:   sdp.QueryError_NOSCOPE,
			ErrorString: fmt.Sprintf("ARN scope %v does not match source scope %v", arnScope, scope),
			Scope:       scope,
		}
	}

	// If the ARN was parsed we can just ask Get for the item
	item, err := getImpl(ctx, client, scope, a.ResourceID())

	if err != nil {
		return nil, sdp.NewQueryError(err)
	}

	return []*sdp.Item{item}, nil
}

// Weight Returns the priority weighting of items returned by this source.
// This is used to resolve conflicts where two sources of the same type
// return an item for a GET request. In this instance only one item can be
// seen on, so the one with the higher weight value will win.
func (s *S3Source) Weight() int {
	return 100
}

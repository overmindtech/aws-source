package ecs

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func (t *TestClient) DescribeClusters(ctx context.Context, params *ecs.DescribeClustersInput, optFns ...func(*ecs.Options)) (*ecs.DescribeClustersOutput, error) {
	return &ecs.DescribeClustersOutput{
		Clusters: []types.Cluster{
			{
				ClusterArn:                        sources.PtrString("arn:aws:ecs:eu-west-2:052392120703:cluster/default"),
				ClusterName:                       sources.PtrString("default"),
				Status:                            sources.PtrString("ACTIVE"),
				RegisteredContainerInstancesCount: 0,
				RunningTasksCount:                 1,
				PendingTasksCount:                 0,
				ActiveServicesCount:               1,
				Statistics: []types.KeyValuePair{
					{
						Name:  sources.PtrString("key"),
						Value: sources.PtrString("value"),
					},
				},
				Tags: []types.Tag{},
				Settings: []types.ClusterSetting{
					{
						Name:  types.ClusterSettingNameContainerInsights,
						Value: sources.PtrString("ENABLED"),
					},
				},
				CapacityProviders: []string{
					"test",
				},
				DefaultCapacityProviderStrategy: []types.CapacityProviderStrategyItem{
					{
						CapacityProvider: sources.PtrString("provider"),
						Base:             10,
						Weight:           100,
					},
				},
				Attachments: []types.Attachment{
					{
						Id:     sources.PtrString("1c1f9cf4-461c-4072-aab2-e2dd346c53e1"),
						Type:   sources.PtrString("as_policy"),
						Status: sources.PtrString("CREATED"),
						Details: []types.KeyValuePair{
							{
								Name:  sources.PtrString("capacityProviderName"),
								Value: sources.PtrString("test"),
							},
							{
								Name:  sources.PtrString("scalingPolicyName"),
								Value: sources.PtrString("ECSManagedAutoScalingPolicy-d2f110eb-20a6-4278-9c1c-47d98e21b1ed"),
							},
						},
					},
				},
				AttachmentsStatus: sources.PtrString("UPDATE_COMPLETE"),
				Configuration: &types.ClusterConfiguration{
					ExecuteCommandConfiguration: &types.ExecuteCommandConfiguration{
						KmsKeyId: sources.PtrString("id"),
						LogConfiguration: &types.ExecuteCommandLogConfiguration{
							CloudWatchEncryptionEnabled: true,
							CloudWatchLogGroupName:      sources.PtrString("cloud-watch-name"),
							S3BucketName:                sources.PtrString("s3-name"),
							S3EncryptionEnabled:         true,
							S3KeyPrefix:                 sources.PtrString("prod"),
						},
					},
				},
				ServiceConnectDefaults: &types.ClusterServiceConnectDefaults{
					Namespace: sources.PtrString("prod"),
				},
			},
		},
	}, nil
}

func (t *TestClient) ListClusters(context.Context, *ecs.ListClustersInput, ...func(*ecs.Options)) (*ecs.ListClustersOutput, error) {
	return &ecs.ListClustersOutput{
		ClusterArns: []string{
			"arn:aws:service:region:account:cluster/name",
		},
	}, nil
}

func TestClusterGetFunc(t *testing.T) {
	item, err := ClusterGetFunc(context.Background(), &TestClient{}, "foo", &ecs.DescribeClustersInput{})

	if err != nil {
		t.Error(err)
	}

	if err = item.Validate(); err != nil {
		t.Error(err)
	}

	tests := sources.ItemRequestTests{
		{
			ExpectedType:   "kms-key",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  "id",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "logs-log-group",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  "cloud-watch-name",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "s3-bucket",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  "s3-name",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "ecs-capacity-provider",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  "test",
			ExpectedScope:  "foo",
		},
	}

	tests.Execute(t, item)
}

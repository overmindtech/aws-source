package cloudwatch

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/overmindtech/aws-source/adapters"
	"github.com/overmindtech/sdp-go"
)

type testCloudwatchClient struct{}

func (c testCloudwatchClient) ListTagsForResource(ctx context.Context, params *cloudwatch.ListTagsForResourceInput, optFns ...func(*cloudwatch.Options)) (*cloudwatch.ListTagsForResourceOutput, error) {
	return &cloudwatch.ListTagsForResourceOutput{
		Tags: []types.Tag{
			{
				Key:   adapters.PtrString("Name"),
				Value: adapters.PtrString("example"),
			},
		},
	}, nil
}

func (c testCloudwatchClient) DescribeAlarms(ctx context.Context, params *cloudwatch.DescribeAlarmsInput, optFns ...func(*cloudwatch.Options)) (*cloudwatch.DescribeAlarmsOutput, error) {
	return nil, nil
}

func (c testCloudwatchClient) DescribeAlarmsForMetric(ctx context.Context, params *cloudwatch.DescribeAlarmsForMetricInput, optFns ...func(*cloudwatch.Options)) (*cloudwatch.DescribeAlarmsForMetricOutput, error) {
	return nil, nil
}

func TestAlarmOutputMapper(t *testing.T) {
	output := &cloudwatch.DescribeAlarmsOutput{
		MetricAlarms: []types.MetricAlarm{
			{
				AlarmName:                          adapters.PtrString("TargetTracking-table/dylan-tfstate-AlarmHigh-14069c4a-6dcc-48a2-bfe6-b5547c90c43d"),
				AlarmArn:                           adapters.PtrString("arn:aws:cloudwatch:eu-west-2:052392120703:alarm:TargetTracking-table/dylan-tfstate-AlarmHigh-14069c4a-6dcc-48a2-bfe6-b5547c90c43d"),
				AlarmDescription:                   adapters.PtrString("DO NOT EDIT OR DELETE. For TargetTrackingScaling policy arn:aws:autoscaling:eu-west-2:052392120703:scalingPolicy:32f3f053-dc75-46fa-9cd4-8e8c34c47b37:resource/dynamodb/table/dylan-tfstate:policyName/$dylan-tfstate-scaling-policy:createdBy/e5bd51d8-94a8-461e-a989-08f4d10b326b."),
				AlarmConfigurationUpdatedTimestamp: adapters.PtrTime(time.Now()),
				ActionsEnabled:                     adapters.PtrBool(true),
				OKActions: []string{
					"arn:aws:autoscaling:eu-west-2:052392120703:scalingPolicy:32f3f053-dc75-46fa-9cd4-8e8c34c47b37:resource/dynamodb/table/dylan-tfstate:policyName/$dylan-tfstate-scaling-policy:createdBy/e5bd51d8-94a8-461e-a989-08f4d10b326b",
				},
				AlarmActions: []string{
					"arn:aws:autoscaling:eu-west-2:052392120703:scalingPolicy:32f3f053-dc75-46fa-9cd4-8e8c34c47b37:resource/dynamodb/table/dylan-tfstate:policyName/$dylan-tfstate-scaling-policy:createdBy/e5bd51d8-94a8-461e-a989-08f4d10b326b",
				},
				InsufficientDataActions: []string{
					"arn:aws:autoscaling:eu-west-2:052392120703:scalingPolicy:32f3f053-dc75-46fa-9cd4-8e8c34c47b37:resource/dynamodb/table/dylan-tfstate:policyName/$dylan-tfstate-scaling-policy:createdBy/e5bd51d8-94a8-461e-a989-08f4d10b326b",
				},
				StateValue:            types.StateValueOk,
				StateReason:           adapters.PtrString("Threshold Crossed: 2 datapoints [0.0 (09/01/23 14:02:00), 1.0 (09/01/23 14:01:00)] were not greater than the threshold (42.0)."),
				StateReasonData:       adapters.PtrString("{\"version\":\"1.0\",\"queryDate\":\"2023-01-09T14:07:25.504+0000\",\"startDate\":\"2023-01-09T14:01:00.000+0000\",\"statistic\":\"Sum\",\"period\":60,\"recentDatapoints\":[1.0,0.0],\"threshold\":42.0,\"evaluatedDatapoints\":[{\"timestamp\":\"2023-01-09T14:02:00.000+0000\",\"sampleCount\":1.0,\"value\":0.0}]}"),
				StateUpdatedTimestamp: adapters.PtrTime(time.Now()),
				MetricName:            adapters.PtrString("ConsumedWriteCapacityUnits"),
				Namespace:             adapters.PtrString("AWS/DynamoDB"),
				Statistic:             types.StatisticSum,
				Dimensions: []types.Dimension{
					{
						Name:  adapters.PtrString("TableName"),
						Value: adapters.PtrString("dylan-tfstate"),
					},
				},
				Period:                     adapters.PtrInt32(60),
				EvaluationPeriods:          adapters.PtrInt32(2),
				Threshold:                  adapters.PtrFloat64(42.0),
				ComparisonOperator:         types.ComparisonOperatorGreaterThanThreshold,
				StateTransitionedTimestamp: adapters.PtrTime(time.Now()),
			},
		},
		CompositeAlarms: []types.CompositeAlarm{
			{
				AlarmName:                          adapters.PtrString("TargetTracking2-table/dylan-tfstate-AlarmHigh-14069c4a-6dcc-48a2-bfe6-b5547c90c43d"),
				AlarmArn:                           adapters.PtrString("arn:aws:cloudwatch:eu-west-2:052392120703:alarm:TargetTracking2-table/dylan-tfstate-AlarmHigh-14069c4a-6dcc-48a2-bfe6-b5547c90c43d"),
				AlarmDescription:                   adapters.PtrString("DO NOT EDIT OR DELETE. For TargetTrackingScaling policy arn:aws:autoscaling:eu-west-2:052392120703:scalingPolicy:32f3f053-dc75-46fa-9cd4-8e8c34c47b37:resource/dynamodb/table/dylan-tfstate:policyName/$dylan-tfstate-scaling-policy:createdBy/e5bd51d8-94a8-461e-a989-08f4d10b326b."),
				AlarmConfigurationUpdatedTimestamp: adapters.PtrTime(time.Now()),
				ActionsEnabled:                     adapters.PtrBool(true),
				OKActions: []string{
					"arn:aws:autoscaling:eu-west-2:052392120703:scalingPolicy:32f3f053-dc75-46fa-9cd4-8e8c34c47b37:resource/dynamodb/table/dylan-tfstate:policyName/$dylan-tfstate-scaling-policy:createdBy/e5bd51d8-94a8-461e-a989-08f4d10b326b",
				},
				AlarmActions: []string{
					"arn:aws:autoscaling:eu-west-2:052392120703:scalingPolicy:32f3f053-dc75-46fa-9cd4-8e8c34c47b37:resource/dynamodb/table/dylan-tfstate:policyName/$dylan-tfstate-scaling-policy:createdBy/e5bd51d8-94a8-461e-a989-08f4d10b326b",
				},
				InsufficientDataActions: []string{
					"arn:aws:autoscaling:eu-west-2:052392120703:scalingPolicy:32f3f053-dc75-46fa-9cd4-8e8c34c47b37:resource/dynamodb/table/dylan-tfstate:policyName/$dylan-tfstate-scaling-policy:createdBy/e5bd51d8-94a8-461e-a989-08f4d10b326b",
				},
				StateValue:                 types.StateValueOk,
				StateReason:                adapters.PtrString("Threshold Crossed: 2 datapoints [0.0 (09/01/23 14:02:00), 1.0 (09/01/23 14:01:00)] were not greater than the threshold (42.0)."),
				StateReasonData:            adapters.PtrString("{\"version\":\"1.0\",\"queryDate\":\"2023-01-09T14:07:25.504+0000\",\"startDate\":\"2023-01-09T14:01:00.000+0000\",\"statistic\":\"Sum\",\"period\":60,\"recentDatapoints\":[1.0,0.0],\"threshold\":42.0,\"evaluatedDatapoints\":[{\"timestamp\":\"2023-01-09T14:02:00.000+0000\",\"sampleCount\":1.0,\"value\":0.0}]}"),
				StateUpdatedTimestamp:      adapters.PtrTime(time.Now()),
				StateTransitionedTimestamp: adapters.PtrTime(time.Now()),
				ActionsSuppressedBy:        types.ActionsSuppressedByAlarm,
				ActionsSuppressedReason:    adapters.PtrString("Alarm is in INSUFFICIENT_DATA state"),
				// link
				ActionsSuppressor:                adapters.PtrString("arn:aws:cloudwatch:eu-west-2:052392120703:alarm:TargetTracking2-table/dylan-tfstate-AlarmHigh-14069c4a-6dcc-48a2-bfe6-b5547c90c43d"),
				ActionsSuppressorExtensionPeriod: adapters.PtrInt32(0),
				ActionsSuppressorWaitPeriod:      adapters.PtrInt32(0),
				AlarmRule:                        adapters.PtrString("ALARM TargetTracking2-table/dylan-tfstate-AlarmHigh-14069c4a-6dcc-48a2-bfe6-b5547c90c43d"),
			},
		},
	}

	scope := "123456789012.eu-west-2"
	items, err := alarmOutputMapper(context.Background(), testCloudwatchClient{}, scope, &cloudwatch.DescribeAlarmsInput{}, output)

	if err != nil {
		t.Error(err)
	}

	if len(items) != 2 {
		t.Fatalf("Expected 2 items, got %d", len(items))
	}

	item := items[1]

	if err = item.Validate(); err != nil {
		t.Error(err)
	}

	if item.GetTags()["Name"] != "example" {
		t.Errorf("Expected tag Name to be example, got %s", item.GetTags()["Name"])
	}

	tests := adapters.QueryTests{
		{
			ExpectedType:   "cloudwatch-alarm",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "TargetTracking2-table/dylan-tfstate-AlarmHigh-14069c4a-6dcc-48a2-bfe6-b5547c90c43d",
			ExpectedScope:  "052392120703.eu-west-2",
		},
	}

	tests.Execute(t, item)

	item = items[0]

	if err = item.Validate(); err != nil {
		t.Error(err)
	}

	tests = adapters.QueryTests{
		{
			ExpectedType:   "dynamodb-table",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "dylan-tfstate",
			ExpectedScope:  scope,
		},
	}

	tests.Execute(t, item)
}

func TestNewAlarmSource(t *testing.T) {
	config, account, region := adapters.GetAutoConfig(t)
	client := cloudwatch.NewFromConfig(config)

	source := NewAlarmSource(client, account, region)

	test := adapters.E2ETest{
		Adapter: source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}

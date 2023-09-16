package cloudwatch

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func TestAlarmOutputMapper(t *testing.T) {
	output := &cloudwatch.DescribeAlarmsOutput{
		MetricAlarms: []types.MetricAlarm{
			{
				AlarmName:                          sources.PtrString("TargetTracking-table/dylan-tfstate-AlarmHigh-14069c4a-6dcc-48a2-bfe6-b5547c90c43d"),
				AlarmArn:                           sources.PtrString("arn:aws:cloudwatch:eu-west-2:052392120703:alarm:TargetTracking-table/dylan-tfstate-AlarmHigh-14069c4a-6dcc-48a2-bfe6-b5547c90c43d"),
				AlarmDescription:                   sources.PtrString("DO NOT EDIT OR DELETE. For TargetTrackingScaling policy arn:aws:autoscaling:eu-west-2:052392120703:scalingPolicy:32f3f053-dc75-46fa-9cd4-8e8c34c47b37:resource/dynamodb/table/dylan-tfstate:policyName/$dylan-tfstate-scaling-policy:createdBy/e5bd51d8-94a8-461e-a989-08f4d10b326b."),
				AlarmConfigurationUpdatedTimestamp: sources.PtrTime(time.Now()),
				ActionsEnabled:                     sources.PtrBool(true),
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
				StateReason:           sources.PtrString("Threshold Crossed: 2 datapoints [0.0 (09/01/23 14:02:00), 1.0 (09/01/23 14:01:00)] were not greater than the threshold (42.0)."),
				StateReasonData:       sources.PtrString("{\"version\":\"1.0\",\"queryDate\":\"2023-01-09T14:07:25.504+0000\",\"startDate\":\"2023-01-09T14:01:00.000+0000\",\"statistic\":\"Sum\",\"period\":60,\"recentDatapoints\":[1.0,0.0],\"threshold\":42.0,\"evaluatedDatapoints\":[{\"timestamp\":\"2023-01-09T14:02:00.000+0000\",\"sampleCount\":1.0,\"value\":0.0}]}"),
				StateUpdatedTimestamp: sources.PtrTime(time.Now()),
				MetricName:            sources.PtrString("ConsumedWriteCapacityUnits"),
				Namespace:             sources.PtrString("AWS/DynamoDB"),
				Statistic:             types.StatisticSum,
				Dimensions: []types.Dimension{
					{
						Name:  sources.PtrString("TableName"),
						Value: sources.PtrString("dylan-tfstate"),
					},
				},
				Period:                     sources.PtrInt32(60),
				EvaluationPeriods:          sources.PtrInt32(2),
				Threshold:                  sources.PtrFloat64(42.0),
				ComparisonOperator:         types.ComparisonOperatorGreaterThanThreshold,
				StateTransitionedTimestamp: sources.PtrTime(time.Now()),
			},
		},
		CompositeAlarms: []types.CompositeAlarm{
			{
				AlarmName:                          sources.PtrString("TargetTracking2-table/dylan-tfstate-AlarmHigh-14069c4a-6dcc-48a2-bfe6-b5547c90c43d"),
				AlarmArn:                           sources.PtrString("arn:aws:cloudwatch:eu-west-2:052392120703:alarm:TargetTracking2-table/dylan-tfstate-AlarmHigh-14069c4a-6dcc-48a2-bfe6-b5547c90c43d"),
				AlarmDescription:                   sources.PtrString("DO NOT EDIT OR DELETE. For TargetTrackingScaling policy arn:aws:autoscaling:eu-west-2:052392120703:scalingPolicy:32f3f053-dc75-46fa-9cd4-8e8c34c47b37:resource/dynamodb/table/dylan-tfstate:policyName/$dylan-tfstate-scaling-policy:createdBy/e5bd51d8-94a8-461e-a989-08f4d10b326b."),
				AlarmConfigurationUpdatedTimestamp: sources.PtrTime(time.Now()),
				ActionsEnabled:                     sources.PtrBool(true),
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
				StateReason:                sources.PtrString("Threshold Crossed: 2 datapoints [0.0 (09/01/23 14:02:00), 1.0 (09/01/23 14:01:00)] were not greater than the threshold (42.0)."),
				StateReasonData:            sources.PtrString("{\"version\":\"1.0\",\"queryDate\":\"2023-01-09T14:07:25.504+0000\",\"startDate\":\"2023-01-09T14:01:00.000+0000\",\"statistic\":\"Sum\",\"period\":60,\"recentDatapoints\":[1.0,0.0],\"threshold\":42.0,\"evaluatedDatapoints\":[{\"timestamp\":\"2023-01-09T14:02:00.000+0000\",\"sampleCount\":1.0,\"value\":0.0}]}"),
				StateUpdatedTimestamp:      sources.PtrTime(time.Now()),
				StateTransitionedTimestamp: sources.PtrTime(time.Now()),
				ActionsSuppressedBy:        types.ActionsSuppressedByAlarm,
				ActionsSuppressedReason:    sources.PtrString("Alarm is in INSUFFICIENT_DATA state"),
				// link
				ActionsSuppressor:                sources.PtrString("arn:aws:cloudwatch:eu-west-2:052392120703:alarm:TargetTracking2-table/dylan-tfstate-AlarmHigh-14069c4a-6dcc-48a2-bfe6-b5547c90c43d"),
				ActionsSuppressorExtensionPeriod: sources.PtrInt32(0),
				ActionsSuppressorWaitPeriod:      sources.PtrInt32(0),
				AlarmRule:                        sources.PtrString("ALARM TargetTracking2-table/dylan-tfstate-AlarmHigh-14069c4a-6dcc-48a2-bfe6-b5547c90c43d"),
			},
		},
	}

	scope := "123456789012.eu-west-2"
	items, err := alarmOutputMapper(scope, &cloudwatch.DescribeAlarmsInput{}, output)

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

	tests := sources.QueryTests{
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

	tests = sources.QueryTests{
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
	config, account, _ := sources.GetAutoConfig(t)

	source := NewAlarmSource(config, account)

	test := sources.E2ETest{
		Source:  source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}

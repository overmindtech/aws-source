package cloudwatch

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

type CloudwatchClient interface {
	ListTagsForResource(ctx context.Context, params *cloudwatch.ListTagsForResourceInput, optFns ...func(*cloudwatch.Options)) (*cloudwatch.ListTagsForResourceOutput, error)
	DescribeAlarms(ctx context.Context, params *cloudwatch.DescribeAlarmsInput, optFns ...func(*cloudwatch.Options)) (*cloudwatch.DescribeAlarmsOutput, error)
	DescribeAlarmsForMetric(ctx context.Context, params *cloudwatch.DescribeAlarmsForMetricInput, optFns ...func(*cloudwatch.Options)) (*cloudwatch.DescribeAlarmsForMetricOutput, error)
}

// ToQueryString Converts an alarm query input to the correct for search string
func ToQueryString(input *cloudwatch.DescribeAlarmsForMetricInput) (string, error) {
	b, err := json.Marshal(input)

	if err != nil {
		return "", err
	}

	return string(b), nil
}

// fromQueryString Converts a search string to an alarm query input
func fromQueryString(query string) (*cloudwatch.DescribeAlarmsForMetricInput, error) {
	input := &cloudwatch.DescribeAlarmsForMetricInput{}

	if err := json.Unmarshal([]byte(query), input); err != nil {
		return nil, err
	}

	return input, nil
}

// Converts cloudwatch tags to a map
func tagsToMap(tags []types.Tag) map[string]string {
	out := make(map[string]string)

	for _, tag := range tags {
		out[*tag.Key] = *tag.Value
	}

	return out
}

type Alarm struct {
	Metric    *types.MetricAlarm
	Composite *types.CompositeAlarm
}

func alarmOutputMapper(ctx context.Context, client CloudwatchClient, scope string, input *cloudwatch.DescribeAlarmsInput, output *cloudwatch.DescribeAlarmsOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	allAlarms := make([]Alarm, 0)

	for i := range output.MetricAlarms {
		allAlarms = append(allAlarms, Alarm{Metric: &output.MetricAlarms[i]})
	}
	for i := range output.CompositeAlarms {
		allAlarms = append(allAlarms, Alarm{Composite: &output.CompositeAlarms[i]})
	}

	for _, alarm := range allAlarms {
		var attrs *sdp.ItemAttributes
		var err error
		var arn *string

		if alarm.Metric != nil {
			attrs, err = sources.ToAttributesCase(alarm.Metric)
			arn = alarm.Metric.AlarmArn
		}
		if alarm.Composite != nil {
			attrs, err = sources.ToAttributesCase(alarm.Composite)
			arn = alarm.Composite.AlarmArn
		}

		if err != nil {
			return nil, err
		}

		var tags map[string]string

		// Get the tags
		tagsOut, err := client.ListTagsForResource(ctx, &cloudwatch.ListTagsForResourceInput{
			ResourceARN: arn,
		})

		if err == nil {
			tags = tagsToMap(tagsOut.Tags)
		} else {
			tags = sources.HandleTagsError(ctx, err)
		}

		if err != nil {
			return nil, err
		}

		item := sdp.Item{
			Type:            "cloudwatch-alarm",
			UniqueAttribute: "alarmName",
			Scope:           scope,
			Attributes:      attrs,
			Tags:            tags,
		}

		// Combine all actions so that we can link the targeted item
		allActions := make([]string, 0)
		if alarm.Metric != nil {
			allActions = append(allActions, alarm.Metric.OKActions...)
			allActions = append(allActions, alarm.Metric.AlarmActions...)
			allActions = append(allActions, alarm.Metric.InsufficientDataActions...)
		}
		if alarm.Composite != nil {
			allActions = append(allActions, alarm.Composite.OKActions...)
			allActions = append(allActions, alarm.Composite.AlarmActions...)
			allActions = append(allActions, alarm.Composite.InsufficientDataActions...)
		}

		for _, action := range allActions {
			if q, err := actionToLink(action); err == nil {
				item.LinkedItemQueries = append(item.LinkedItemQueries, q)
			}
		}

		// Calculate state and convert this to health
		var stateValue types.StateValue
		if alarm.Metric != nil {
			stateValue = alarm.Metric.StateValue
		}
		if alarm.Composite != nil {
			stateValue = alarm.Composite.StateValue
		}

		switch stateValue {
		case types.StateValueOk:
			item.Health = sdp.Health_HEALTH_OK.Enum()
		case types.StateValueAlarm:
			item.Health = sdp.Health_HEALTH_ERROR.Enum()
		case types.StateValueInsufficientData:
			item.Health = sdp.Health_HEALTH_UNKNOWN.Enum()
		}

		// Link to the suppressor alarm
		if alarm.Composite != nil && alarm.Composite.ActionsSuppressor != nil {
			if arn, err := sources.ParseARN(*alarm.Composite.ActionsSuppressor); err == nil {
				// +overmind:link cloudwatch-alarm
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "cloudwatch-alarm",
						Method: sdp.QueryMethod_GET,
						Query:  arn.ResourceID(),
						Scope:  sources.FormatScope(arn.AccountID, arn.Region),
					},
					BlastPropagation: &sdp.BlastPropagation{
						// Changes to the suppressor alarm will affect this alarm
						In: true,
						// Changes to this alarm won't affect the suppressor alarm
						Out: false,
					},
				})
			}
		}

		if alarm.Metric != nil && alarm.Metric.Namespace != nil {
			// Possible links for a metric alarm
			//
			// +overmind:link acm-certificate
			// +overmind:link autoscaling-auto-scaling-group
			// +overmind:link backup-backup-vault
			// +overmind:link dynamodb-table
			// +overmind:link ec2-image
			// +overmind:link ec2-instance
			// +overmind:link ec2-nat-gateway
			// +overmind:link ec2-volume
			// +overmind:link ecs-cluster
			// +overmind:link ecs-service
			// +overmind:link efs-file-system
			// +overmind:link elb-load-balancer
			// +overmind:link elbv2-load-balancer
			// +overmind:link elbv2-target-group
			// +overmind:link lambda-function
			// +overmind:link rds-db-cluster
			// +overmind:link rds-db-instance
			// +overmind:link route53-health-check
			// +overmind:link route53-hosted-zone
			// +overmind:link s3-bucket

			// Check for links based on the metric that is being monitored
			q, err := SuggestedQuery(*alarm.Metric.Namespace, scope, alarm.Metric.Dimensions)

			if err == nil {
				item.LinkedItemQueries = append(item.LinkedItemQueries, q)
			}
		}

		items = append(items, &item)
	}

	return items, nil
}

//go:generate docgen ../../docs-data
// +overmind:type cloudwatch-alarm
// +overmind:descriptiveType CloudWatch Alarm
// +overmind:get Get an alarm by name
// +overmind:list List all alarms
// +overmind:search Search for alarms. This accepts JSON in the format of `cloudwatch.DescribeAlarmsForMetricInput`
// +overmind:group AWS
// +overmind:terraform:queryMap aws_cloudwatch_metric_alarm.alarm_name

func NewAlarmSource(client *cloudwatch.Client, accountID string, region string) *sources.DescribeOnlySource[*cloudwatch.DescribeAlarmsInput, *cloudwatch.DescribeAlarmsOutput, CloudwatchClient, *cloudwatch.Options] {
	return &sources.DescribeOnlySource[*cloudwatch.DescribeAlarmsInput, *cloudwatch.DescribeAlarmsOutput, CloudwatchClient, *cloudwatch.Options]{
		ItemType:  "cloudwatch-alarm",
		Client:    client,
		Region:    region,
		AccountID: accountID,
		PaginatorBuilder: func(client CloudwatchClient, params *cloudwatch.DescribeAlarmsInput) sources.Paginator[*cloudwatch.DescribeAlarmsOutput, *cloudwatch.Options] {
			return cloudwatch.NewDescribeAlarmsPaginator(client, params)
		},
		DescribeFunc: func(ctx context.Context, client CloudwatchClient, input *cloudwatch.DescribeAlarmsInput) (*cloudwatch.DescribeAlarmsOutput, error) {
			return client.DescribeAlarms(ctx, input)
		},
		InputMapperGet: func(scope, query string) (*cloudwatch.DescribeAlarmsInput, error) {
			return &cloudwatch.DescribeAlarmsInput{
				AlarmNames: []string{query},
			}, nil
		},
		InputMapperList: func(scope string) (*cloudwatch.DescribeAlarmsInput, error) {
			return &cloudwatch.DescribeAlarmsInput{}, nil
		},
		InputMapperSearch: func(ctx context.Context, client CloudwatchClient, scope, query string) (*cloudwatch.DescribeAlarmsInput, error) {
			// Search uses the DescribeAlarmsForMetric API call to find alarms
			// based on a JSON input
			input, err := fromQueryString(query)

			if err != nil {
				return nil, err
			}

			out, err := client.DescribeAlarmsForMetric(ctx, input)

			if err != nil {
				return nil, err
			}

			name := make([]string, 0)

			for _, alarm := range out.MetricAlarms {
				if alarm.AlarmName != nil {
					name = append(name, *alarm.AlarmName)
				}
			}

			return &cloudwatch.DescribeAlarmsInput{
				AlarmNames: name,
			}, nil
		},
		OutputMapper: alarmOutputMapper,
	}
}

// actionToLink converts an action string to a link to the resource that the
// action refers to. The actions to execute when this alarm transitions to the
// ALARM state from any other state. Each action is specified as an Amazon
// Resource Name (ARN). Valid values: EC2 actions:
//
// * arn:aws:automate:region:ec2:stop
//
// * arn:aws:automate:region:ec2:terminate
//
// * arn:aws:automate:region:ec2:reboot
//
// * arn:aws:automate:region:ec2:recover
//
// * arn:aws:swf:region:account-id:action/actions/AWS_EC2.InstanceId.Stop/1.0
//
// *
// arn:aws:swf:region:account-id:action/actions/AWS_EC2.InstanceId.Terminate/1.0
//
// * arn:aws:swf:region:account-id:action/actions/AWS_EC2.InstanceId.Reboot/1.0
//
// * arn:aws:swf:region:account-id:action/actions/AWS_EC2.InstanceId.Recover/1.0
//
// Autoscaling action:
//
// *
// arn:aws:autoscaling:region:account-id:scalingPolicy:policy-id:autoScalingGroupName/group-friendly-name:policyName/policy-friendly-name
//
// SSN notification action:
//
// *
// arn:aws:sns:region:account-id:sns-topic-name:autoScalingGroupName/group-friendly-name:policyName/policy-friendly-name
//
// SSM integration actions:
//
// * arn:aws:ssm:region:account-id:opsitem:severity#CATEGORY=category-name
//
// * arn:aws:ssm-incidents::account-id:responseplan/response-plan-name
func actionToLink(action string) (*sdp.LinkedItemQuery, error) {
	arn, err := sources.ParseARN(action)

	if err != nil {
		return nil, err
	}

	switch arn.Service {
	case "autoscaling":
		// +overmind:link autoscaling-policy
		return &sdp.LinkedItemQuery{
			Query: &sdp.Query{
				Type:   "autoscaling-policy",
				Method: sdp.QueryMethod_SEARCH,
				Query:  action,
				Scope:  sources.FormatScope(arn.AccountID, arn.Region),
			},
			BlastPropagation: &sdp.BlastPropagation{
				// Changes to the policy won't affect the alarm
				In: false,
				// Changes to the metric alarm will affect the policy
				Out: true,
			},
		}, nil
	case "sns":
		// +overmind:link sns-topic
		return &sdp.LinkedItemQuery{
			Query: &sdp.Query{
				Type:   "sns-topic",
				Method: sdp.QueryMethod_SEARCH,
				Query:  action,
				Scope:  sources.FormatScope(arn.AccountID, arn.Region),
			},
			BlastPropagation: &sdp.BlastPropagation{
				// Changes to the topic won't affect the alarm
				In: false,
				// Changes to the alarm will affect the topic
				Out: true,
			},
		}, nil
	case "ssm":
		// +overmind:link ssm-ops-item
		return &sdp.LinkedItemQuery{
			Query: &sdp.Query{
				Type:   "ssm-ops-item",
				Method: sdp.QueryMethod_SEARCH,
				Query:  action,
				Scope:  sources.FormatScope(arn.AccountID, arn.Region),
			},
			BlastPropagation: &sdp.BlastPropagation{
				// Changes to an ops item won't affect the alarm
				In: false,
				// Changes to the alarm will affect the ops item
				Out: true,
			},
		}, nil
	case "ssm-incidents":
		// +overmind:link ssm-incidents-response-plan
		return &sdp.LinkedItemQuery{
			Query: &sdp.Query{
				Type:   "ssm-incidents-response-plan",
				Method: sdp.QueryMethod_SEARCH,
				Query:  action,
				Scope:  sources.FormatScope(arn.AccountID, arn.Region),
			},
			BlastPropagation: &sdp.BlastPropagation{
				// Changes to a response plan won't affect the alarm
				In: false,
				// Changes to the alarm will affect the response plan
				Out: true,
			},
		}, nil
	default:
		return nil, errors.New("unknown service in ARN: " + arn.Service)
	}
}

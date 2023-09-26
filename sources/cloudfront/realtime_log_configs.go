package cloudfront

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func realtimeLogConfigsItemMapper(scope string, awsItem *types.RealtimeLogConfig) (*sdp.Item, error) {
	attributes, err := sources.ToAttributesCase(awsItem)

	if err != nil {
		return nil, err
	}

	item := sdp.Item{
		Type:            "cloudfront-realtime-log-config",
		UniqueAttribute: "name",
		Attributes:      attributes,
		Scope:           scope,
	}

	for _, endpoint := range awsItem.EndPoints {
		if endpoint.KinesisStreamConfig != nil {
			if endpoint.KinesisStreamConfig.RoleARN != nil {
				if arn, err := sources.ParseARN(*endpoint.KinesisStreamConfig.RoleARN); err == nil {
					// +overmind:link iam-role
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
						Query: &sdp.Query{
							Type:   "iam-role",
							Method: sdp.QueryMethod_SEARCH,
							Query:  *endpoint.KinesisStreamConfig.RoleARN,
							Scope:  sources.FormatScope(arn.AccountID, arn.Region),
						},
						BlastPropagation: &sdp.BlastPropagation{
							// Changes to the role will affect us
							In: true,
							// We can't affect the role
							Out: false,
						},
					})
				}
			}

			if endpoint.KinesisStreamConfig.StreamARN != nil {
				if arn, err := sources.ParseARN(*endpoint.KinesisStreamConfig.StreamARN); err == nil {
					// +overmind:link kinesis-stream
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
						Query: &sdp.Query{
							Type:   "kinesis-stream",
							Method: sdp.QueryMethod_SEARCH,
							Query:  *endpoint.KinesisStreamConfig.StreamARN,
							Scope:  sources.FormatScope(arn.AccountID, arn.Region),
						},
						BlastPropagation: &sdp.BlastPropagation{
							// Changes to this will affect the stream
							Out: true,
							// The stream can affect us
							In: true,
						},
					})
				}
			}
		}
	}

	return &item, nil
}

//go:generate docgen ../../docs-data
// +overmind:type cloudfront-realtime-log-config
// +overmind:descriptiveType CloudFront Realtime Log Config
// +overmind:get Get Realtime Log Config by Name
// +overmind:list List Realtime Log Configs
// +overmind:search Search Realtime Log Configs by ARN
// +overmind:group AWS
// +overmind:terraform:queryMap aws_cloudfront_realtime_log_config.arn
// +overmind:terraform:method SEARCH

func NewRealtimeLogConfigsSource(config aws.Config, accountID string) *sources.GetListSource[*types.RealtimeLogConfig, *cloudfront.Client, *cloudfront.Options] {
	return &sources.GetListSource[*types.RealtimeLogConfig, *cloudfront.Client, *cloudfront.Options]{
		ItemType:  "cloudfront-realtime-log-config",
		Client:    cloudfront.NewFromConfig(config),
		AccountID: accountID,
		Region:    "", // Cloudfront resources aren't tied to a region
		GetFunc: func(ctx context.Context, client *cloudfront.Client, scope, query string) (*types.RealtimeLogConfig, error) {
			out, err := client.GetRealtimeLogConfig(ctx, &cloudfront.GetRealtimeLogConfigInput{
				Name: &query,
			})

			if err != nil {
				return nil, err
			}

			return out.RealtimeLogConfig, nil
		},
		ListFunc: func(ctx context.Context, client *cloudfront.Client, scope string) ([]*types.RealtimeLogConfig, error) {
			out, err := client.ListRealtimeLogConfigs(ctx, &cloudfront.ListRealtimeLogConfigsInput{})

			if err != nil {
				return nil, err
			}

			logConfigs := make([]*types.RealtimeLogConfig, len(out.RealtimeLogConfigs.Items))

			for i, logConfig := range out.RealtimeLogConfigs.Items {
				logConfigs[i] = &logConfig
			}

			return logConfigs, nil
		},
		ItemMapper: realtimeLogConfigsItemMapper,
	}
}

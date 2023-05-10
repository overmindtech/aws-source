package route53

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	cwtypes "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/overmindtech/aws-source/sources"
	cw "github.com/overmindtech/aws-source/sources/cloudwatch"
	"github.com/overmindtech/sdp-go"
)

type HealthCheck struct {
	types.HealthCheck
	HealthCheckObservations []types.HealthCheckObservation
}

func healthCheckGetFunc(ctx context.Context, client *route53.Client, scope, query string) (*HealthCheck, error) {
	out, err := client.GetHealthCheck(ctx, &route53.GetHealthCheckInput{
		HealthCheckId: &query,
	})

	if err != nil {
		return nil, err
	}

	status, err := client.GetHealthCheckStatus(ctx, &route53.GetHealthCheckStatusInput{
		HealthCheckId: &query,
	})

	if err != nil {
		return nil, err
	}

	return &HealthCheck{
		HealthCheck:             *out.HealthCheck,
		HealthCheckObservations: status.HealthCheckObservations,
	}, nil
}

func healthCheckListFunc(ctx context.Context, client *route53.Client, scope string) ([]*HealthCheck, error) {
	out, err := client.ListHealthChecks(ctx, &route53.ListHealthChecksInput{})

	if err != nil {
		return nil, err
	}

	healthChecks := make([]*HealthCheck, len(out.HealthChecks))

	for i, healthCheck := range out.HealthChecks {
		status, err := client.GetHealthCheckStatus(ctx, &route53.GetHealthCheckStatusInput{
			HealthCheckId: healthCheck.Id,
		})

		if err != nil {
			return nil, err
		}

		healthChecks[i] = &HealthCheck{
			HealthCheck:             healthCheck,
			HealthCheckObservations: status.HealthCheckObservations,
		}
	}

	return healthChecks, nil
}

func healthCheckItemMapper(scope string, awsItem *HealthCheck) (*sdp.Item, error) {
	attributes, err := sources.ToAttributesCase(awsItem)

	if err != nil {
		return nil, err
	}

	item := sdp.Item{
		Type:            "route53-health-check",
		UniqueAttribute: "id",
		Attributes:      attributes,
		Scope:           scope,
	}

	// Link to the cloudwatch metric that tracks this health check
	query, err := cw.ToQueryString(&cloudwatch.DescribeAlarmsForMetricInput{
		Namespace:  aws.String("AWS/Route53"),
		MetricName: aws.String("HealthCheckStatus"),
		Dimensions: []cwtypes.Dimension{
			{
				Name:  aws.String("HealthCheckId"),
				Value: awsItem.Id,
			},
		},
	})

	if err == nil {
		// +overmind:link cloudwatch-alarm
		item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{Query: &sdp.Query{
			Type:   "cloudwatch-alarm",
			Query:  query,
			Method: sdp.QueryMethod_SEARCH,
			Scope:  scope,
		}})
	}

	healthy := true

	for _, observation := range awsItem.HealthCheckObservations {
		if observation.StatusReport != nil && observation.StatusReport.Status != nil {
			if strings.HasPrefix(*observation.StatusReport.Status, "Failure") {
				healthy = false
			}
		}
	}

	if healthy {
		item.Health = sdp.Health_HEALTH_OK.Enum()
	} else {
		item.Health = sdp.Health_HEALTH_ERROR.Enum()
	}

	return &item, nil
}

//go:generate docgen ../../docs-data
// +overmind:type route53-health-check
// +overmind:descriptiveType Route53 Health Check
// +overmind:get Get health check by ID
// +overmind:list List all health checks
// +overmind:search Search for health checks by ARN
// +overmind:group AWS

func NewHealthCheckSource(config aws.Config, accountID string, region string) *sources.GetListSource[*HealthCheck, *route53.Client, *route53.Options] {
	return &sources.GetListSource[*HealthCheck, *route53.Client, *route53.Options]{
		ItemType:   "route53-health-check",
		Client:     route53.NewFromConfig(config),
		AccountID:  accountID,
		Region:     region,
		GetFunc:    healthCheckGetFunc,
		ListFunc:   healthCheckListFunc,
		ItemMapper: healthCheckItemMapper,
	}
}

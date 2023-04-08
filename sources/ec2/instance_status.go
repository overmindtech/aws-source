package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func instanceStatusInputMapperGet(scope, query string) (*ec2.DescribeInstanceStatusInput, error) {
	return &ec2.DescribeInstanceStatusInput{
		InstanceIds: []string{
			query,
		},
	}, nil
}

func instanceStatusInputMapperList(scope string) (*ec2.DescribeInstanceStatusInput, error) {
	return &ec2.DescribeInstanceStatusInput{}, nil
}

func instanceStatusOutputMapper(scope string, _ *ec2.DescribeInstanceStatusInput, output *ec2.DescribeInstanceStatusOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, instanceStatus := range output.InstanceStatuses {
		attrs, err := sources.ToAttributesCase(instanceStatus)

		if err != nil {
			return nil, &sdp.QueryError{
				ErrorType:   sdp.QueryError_OTHER,
				ErrorString: err.Error(),
				Scope:       scope,
			}
		}

		item := sdp.Item{
			Type:            "ec2-instance-status",
			UniqueAttribute: "instanceId",
			Scope:           scope,
			Attributes:      attrs,
			LinkedItemQueries: []*sdp.Query{
				{
					Type:   "ec2-instance",
					Method: sdp.QueryMethod_GET,
					Query:  *instanceStatus.InstanceId,
					Scope:  scope,
				},
			},
		}

		switch instanceStatus.SystemStatus.Status {
		case types.SummaryStatusOk:
			item.Health = sdp.Health_HEALTH_OK.Enum()
		case types.SummaryStatusImpaired:
			item.Health = sdp.Health_HEALTH_ERROR.Enum()
		case types.SummaryStatusInsufficientData:
			item.Health = sdp.Health_HEALTH_UNKNOWN.Enum()
		case types.SummaryStatusNotApplicable:
			item.Health = nil
		case types.SummaryStatusInitializing:
			// TODO: Decide if I need this
			item.Health = nil
		}

		if instanceStatus.AvailabilityZone != nil {
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
				Type:   "ec2-availability-zone",
				Method: sdp.QueryMethod_GET,
				Query:  *instanceStatus.AvailabilityZone,
				Scope:  scope,
			})
		}

		items = append(items, &item)
	}

	return items, nil
}

func NewInstanceStatusSource(config aws.Config, accountID string, limit *LimitBucket) *sources.DescribeOnlySource[*ec2.DescribeInstanceStatusInput, *ec2.DescribeInstanceStatusOutput, *ec2.Client, *ec2.Options] {
	return &sources.DescribeOnlySource[*ec2.DescribeInstanceStatusInput, *ec2.DescribeInstanceStatusOutput, *ec2.Client, *ec2.Options]{
		Config:    config,
		Client:    ec2.NewFromConfig(config),
		AccountID: accountID,
		ItemType:  "ec2-instance-status",
		DescribeFunc: func(ctx context.Context, client *ec2.Client, input *ec2.DescribeInstanceStatusInput) (*ec2.DescribeInstanceStatusOutput, error) {
			<-limit.C // Wait for late limiting
			return client.DescribeInstanceStatus(ctx, input)
		},
		InputMapperGet:  instanceStatusInputMapperGet,
		InputMapperList: instanceStatusInputMapperList,
		PaginatorBuilder: func(client *ec2.Client, params *ec2.DescribeInstanceStatusInput) sources.Paginator[*ec2.DescribeInstanceStatusOutput, *ec2.Options] {
			return ec2.NewDescribeInstanceStatusPaginator(client, params)
		},
		OutputMapper: instanceStatusOutputMapper,
	}
}

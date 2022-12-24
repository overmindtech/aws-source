package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func InstanceStatusInputMapperGet(scope, query string) (*ec2.DescribeInstanceStatusInput, error) {
	return &ec2.DescribeInstanceStatusInput{
		InstanceIds: []string{
			query,
		},
	}, nil
}

func InstanceStatusInputMapperList(scope string) (*ec2.DescribeInstanceStatusInput, error) {
	return &ec2.DescribeInstanceStatusInput{}, nil
}

func InstanceStatusOutputMapper(scope string, output *ec2.DescribeInstanceStatusOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, instanceStatus := range output.InstanceStatuses {
		attrs, err := sources.ToAttributesCase(instanceStatus)

		if err != nil {
			return nil, &sdp.ItemRequestError{
				ErrorType:   sdp.ItemRequestError_OTHER,
				ErrorString: err.Error(),
				Scope:       scope,
			}
		}

		item := sdp.Item{
			Type:            "ec2-instance-status",
			UniqueAttribute: "instanceId",
			Scope:           scope,
			Attributes:      attrs,
		}

		if instanceStatus.AvailabilityZone != nil {
			item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
				Type:   "ec2-availability-zone",
				Method: sdp.RequestMethod_GET,
				Query:  *instanceStatus.AvailabilityZone,
				Scope:  scope,
			})
		}

		items = append(items, &item)
	}

	return items, nil
}

func NewInstanceStatusSource(config aws.Config, accountID string) *sources.AWSSource[*ec2.DescribeInstanceStatusInput, *ec2.DescribeInstanceStatusOutput, *ec2.Client, *ec2.Options] {
	return &sources.AWSSource[*ec2.DescribeInstanceStatusInput, *ec2.DescribeInstanceStatusOutput, *ec2.Client, *ec2.Options]{
		Config:    config,
		AccountID: accountID,
		ItemType:  "ec2-instance-status",
		DescribeFunc: func(ctx context.Context, client *ec2.Client, input *ec2.DescribeInstanceStatusInput) (*ec2.DescribeInstanceStatusOutput, error) {
			return client.DescribeInstanceStatus(ctx, input)
		},
		InputMapperGet:  InstanceStatusInputMapperGet,
		InputMapperList: InstanceStatusInputMapperList,
		PaginatorBuilder: func(client *ec2.Client, params *ec2.DescribeInstanceStatusInput) sources.Paginator[*ec2.DescribeInstanceStatusOutput, *ec2.Options] {
			return ec2.NewDescribeInstanceStatusPaginator(client, params)
		},
		OutputMapper: InstanceStatusOutputMapper,
	}
}

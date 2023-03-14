package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func reservedInstanceInputMapperGet(scope, query string) (*ec2.DescribeReservedInstancesInput, error) {
	return &ec2.DescribeReservedInstancesInput{
		ReservedInstancesIds: []string{
			query,
		},
	}, nil
}

func reservedInstanceInputMapperList(scope string) (*ec2.DescribeReservedInstancesInput, error) {
	return &ec2.DescribeReservedInstancesInput{}, nil
}

func reservedInstanceOutputMapper(scope string, _ *ec2.DescribeReservedInstancesInput, output *ec2.DescribeReservedInstancesOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, reservation := range output.ReservedInstances {
		attrs, err := sources.ToAttributesCase(reservation)

		if err != nil {
			return nil, &sdp.QueryError{
				ErrorType:   sdp.QueryError_OTHER,
				ErrorString: err.Error(),
				Scope:       scope,
			}
		}

		item := sdp.Item{
			Type:            "ec2-reserved-instance",
			UniqueAttribute: "reservedInstancesId",
			Scope:           scope,
			Attributes:      attrs,
		}

		if reservation.AvailabilityZone != nil {
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
				Type:   "ec2-availability-zone",
				Method: sdp.QueryMethod_GET,
				Query:  *reservation.AvailabilityZone,
				Scope:  scope,
			})
		}

		items = append(items, &item)
	}

	return items, nil
}

func NewReservedInstanceSource(config aws.Config, accountID string, limit *LimitBucket) *sources.DescribeOnlySource[*ec2.DescribeReservedInstancesInput, *ec2.DescribeReservedInstancesOutput, *ec2.Client, *ec2.Options] {
	return &sources.DescribeOnlySource[*ec2.DescribeReservedInstancesInput, *ec2.DescribeReservedInstancesOutput, *ec2.Client, *ec2.Options]{
		Config:    config,
		Client:    ec2.NewFromConfig(config),
		AccountID: accountID,
		ItemType:  "ec2-reserved-instance",
		DescribeFunc: func(ctx context.Context, client *ec2.Client, input *ec2.DescribeReservedInstancesInput) (*ec2.DescribeReservedInstancesOutput, error) {
			<-limit.C // Wait for late limiting
			return client.DescribeReservedInstances(ctx, input)
		},
		InputMapperGet:  reservedInstanceInputMapperGet,
		InputMapperList: reservedInstanceInputMapperList,
		OutputMapper:    reservedInstanceOutputMapper,
	}
}

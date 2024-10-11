package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/overmindtech/aws-source/adapters"
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

func reservedInstanceOutputMapper(_ context.Context, _ *ec2.Client, scope string, _ *ec2.DescribeReservedInstancesInput, output *ec2.DescribeReservedInstancesOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, reservation := range output.ReservedInstances {
		attrs, err := adapters.ToAttributesWithExclude(reservation, "tags")

		if err != nil {
			return nil, &sdp.QueryError{
				ErrorType:   sdp.QueryError_OTHER,
				ErrorString: err.Error(),
				Scope:       scope,
			}
		}

		item := sdp.Item{
			Type:            "ec2-reserved-instance",
			UniqueAttribute: "ReservedInstancesId",
			Scope:           scope,
			Attributes:      attrs,
			Tags:            tagsToMap(reservation.Tags),
		}

		items = append(items, &item)
	}

	return items, nil
}

//go:generate docgen ../../docs-data
// +overmind:type ec2-reserved-instance
// +overmind:descriptiveType Reserved EC2 Instance
// +overmind:get Get a reserved EC2 instance by ID
// +overmind:list List all reserved EC2 instances
// +overmind:search Search reserved EC2 instances by ARN
// +overmind:group AWS

func NewReservedInstanceAdapter(client *ec2.Client, accountID string, region string) *adapters.DescribeOnlyAdapter[*ec2.DescribeReservedInstancesInput, *ec2.DescribeReservedInstancesOutput, *ec2.Client, *ec2.Options] {
	return &adapters.DescribeOnlyAdapter[*ec2.DescribeReservedInstancesInput, *ec2.DescribeReservedInstancesOutput, *ec2.Client, *ec2.Options]{
		Region:          region,
		Client:          client,
		AccountID:       accountID,
		ItemType:        "ec2-reserved-instance",
		AdapterMetadata: ReservedInstanceMetadata(),
		DescribeFunc: func(ctx context.Context, client *ec2.Client, input *ec2.DescribeReservedInstancesInput) (*ec2.DescribeReservedInstancesOutput, error) {
			return client.DescribeReservedInstances(ctx, input)
		},
		InputMapperGet:  reservedInstanceInputMapperGet,
		InputMapperList: reservedInstanceInputMapperList,
		OutputMapper:    reservedInstanceOutputMapper,
	}
}

func ReservedInstanceMetadata() sdp.AdapterMetadata {
	return sdp.AdapterMetadata{
		Type:            "ec2-reserved-instance",
		DescriptiveName: "Reserved EC2 Instance",
		SupportedQueryMethods: &sdp.AdapterSupportedQueryMethods{
			Get:               true,
			List:              true,
			Search:            true,
			GetDescription:    "Get a reserved EC2 instance by ID",
			ListDescription:   "List all reserved EC2 instances",
			SearchDescription: "Search reserved EC2 instances by ARN",
		},
		Category: sdp.AdapterCategory_ADAPTER_CATEGORY_COMPUTE_APPLICATION,
	}
}

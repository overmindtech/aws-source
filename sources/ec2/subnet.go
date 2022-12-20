package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func SubnetInputMapperGet(scope string, query string) (*ec2.DescribeSubnetsInput, error) {
	return &ec2.DescribeSubnetsInput{
		SubnetIds: []string{
			query,
		},
	}, nil
}

func SubnetInputMapperList(scope string) (*ec2.DescribeSubnetsInput, error) {
	return &ec2.DescribeSubnetsInput{}, nil
}

func SubnetOutputMapper(scope string, output *ec2.DescribeSubnetsOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, subnet := range output.Subnets {
		var err error
		var attrs *sdp.ItemAttributes
		attrs, err = sources.ToAttributesCase(subnet)

		if err != nil {
			return nil, &sdp.ItemRequestError{
				ErrorType:   sdp.ItemRequestError_OTHER,
				ErrorString: err.Error(),
				Scope:       scope,
			}
		}

		item := sdp.Item{
			Type:            "ec2-subnet",
			UniqueAttribute: "subnetId",
			Scope:           scope,
			Attributes:      attrs,
		}

		if subnet.AvailabilityZone != nil {
			item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
				Type:   "ec2-availability-zone",
				Method: sdp.RequestMethod_GET,
				Query:  *subnet.AvailabilityZone,
				Scope:  scope,
			})
		}

		if subnet.VpcId != nil {
			item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
				Type:   "ec2-vpc",
				Method: sdp.RequestMethod_GET,
				Query:  *subnet.VpcId,
				Scope:  scope,
			})
		}

		items = append(items, &item)
	}

	return items, nil
}

func NewSubnetSource(config aws.Config, accountID string) *EC2Source[*ec2.DescribeSubnetsInput, *ec2.DescribeSubnetsOutput] {
	return &EC2Source[*ec2.DescribeSubnetsInput, *ec2.DescribeSubnetsOutput]{
		Config:    config,
		AccountID: accountID,
		ItemType:  "ec2-subnet",
		DescribeFunc: func(ctx context.Context, client *ec2.Client, input *ec2.DescribeSubnetsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSubnetsOutput, error) {
			return client.DescribeSubnets(ctx, input)
		},
		InputMapperGet:  SubnetInputMapperGet,
		InputMapperList: SubnetInputMapperList,
		PaginatorBuilder: func(client *ec2.Client, params *ec2.DescribeSubnetsInput) Paginator[*ec2.DescribeSubnetsOutput] {
			return ec2.NewDescribeSubnetsPaginator(client, params)
		},
		OutputMapper: SubnetOutputMapper,
	}
}

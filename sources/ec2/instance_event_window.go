package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func InstanceEventWindowInputMapperGet(scope, query string) (*ec2.DescribeInstanceEventWindowsInput, error) {
	return &ec2.DescribeInstanceEventWindowsInput{
		InstanceEventWindowIds: []string{
			query,
		},
	}, nil
}

func InstanceEventWindowInputMapperList(scope string) (*ec2.DescribeInstanceEventWindowsInput, error) {
	return &ec2.DescribeInstanceEventWindowsInput{}, nil
}

func InstanceEventWindowOutputMapper(scope string, output *ec2.DescribeInstanceEventWindowsOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, ew := range output.InstanceEventWindows {
		attrs, err := sources.ToAttributesCase(ew)

		if err != nil {
			return nil, &sdp.ItemRequestError{
				ErrorType:   sdp.ItemRequestError_OTHER,
				ErrorString: err.Error(),
				Scope:       scope,
			}
		}

		item := sdp.Item{
			Type:            "ec2-instance-event-window",
			UniqueAttribute: "instanceEventWindowId",
			Scope:           scope,
			Attributes:      attrs,
		}

		if at := ew.AssociationTarget; at != nil {
			for _, id := range at.DedicatedHostIds {
				item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
					Type:   "ec2-host",
					Method: sdp.RequestMethod_GET,
					Query:  id,
					Scope:  scope,
				})
			}

			for _, id := range at.InstanceIds {
				item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
					Type:   "ec2-instance",
					Method: sdp.RequestMethod_GET,
					Query:  id,
					Scope:  scope,
				})
			}
		}

		items = append(items, &item)
	}

	return items, nil
}

func NewInstanceEventWindowSource(config aws.Config, accountID string, limit *LimitBucket) *sources.DescribeOnlySource[*ec2.DescribeInstanceEventWindowsInput, *ec2.DescribeInstanceEventWindowsOutput, *ec2.Client, *ec2.Options] {
	return &sources.DescribeOnlySource[*ec2.DescribeInstanceEventWindowsInput, *ec2.DescribeInstanceEventWindowsOutput, *ec2.Client, *ec2.Options]{
		Config:    config,
		AccountID: accountID,
		ItemType:  "ec2-instance-event-window",
		DescribeFunc: func(ctx context.Context, client *ec2.Client, input *ec2.DescribeInstanceEventWindowsInput) (*ec2.DescribeInstanceEventWindowsOutput, error) {
			<-limit.C // Wait for late limiting
			return client.DescribeInstanceEventWindows(ctx, input)
		},
		InputMapperGet:  InstanceEventWindowInputMapperGet,
		InputMapperList: InstanceEventWindowInputMapperList,
		PaginatorBuilder: func(client *ec2.Client, params *ec2.DescribeInstanceEventWindowsInput) sources.Paginator[*ec2.DescribeInstanceEventWindowsOutput, *ec2.Options] {
			return ec2.NewDescribeInstanceEventWindowsPaginator(client, params)
		},
		OutputMapper: InstanceEventWindowOutputMapper,
	}
}

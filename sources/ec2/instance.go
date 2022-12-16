package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/overmindtech/sdp-go"
)

func InstanceInputMapper(scope, query string, method sdp.RequestMethod) (*ec2.DescribeInstancesInput, error) {
	switch method {
	case sdp.RequestMethod_GET:
		return &ec2.DescribeInstancesInput{
			InstanceIds: []string{
				query,
			},
		}, nil
	case sdp.RequestMethod_LIST:
		return &ec2.DescribeInstancesInput{}
	}
}

func InstanceOutputMapper(scope string, output *ec2.DescribeInstancesOutput) ([]*sdp.Item, error) {
}

func NewInstanceSource(config aws.Config, accountID string) *EC2Source[*ec2.DescribeInstancesInput, *ec2.DescribeInstancesOutput] {
	return &EC2Source[*ec2.DescribeInstancesInput, *ec2.DescribeInstancesOutput]{
		Config:    config,
		AccountID: accountID,
		ItemType:  "ec2-instance",
		DescribeFunc: func(ctx context.Context, client *ec2.Client, input *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
			return client.DescribeInstances(ctx, input)
		},
		InputMapper:  InstanceInputMapper,
		OutputMapper: InstanceOutputMapper,
	}
}

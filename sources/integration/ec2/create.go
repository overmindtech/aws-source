package ec2

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/overmindtech/aws-source/sources/integration"
)

func createEC2Instance(ctx context.Context, logger *slog.Logger, client *ec2.Client, testID string) error {
	// check if a resource with the same tags already exists
	id, err := findActiveInstanceIDByTags(client)
	if err != nil {
		if errors.As(err, new(integration.NotFoundError)) {
			logger.InfoContext(ctx, "Creating EC2 instance")
		} else {
			return err
		}
	}

	if id != nil {
		logger.InfoContext(ctx, "EC2 instance already exists")
		return nil
	}

	input := &ec2.RunInstancesInput{
		DryRun: aws.Bool(false),
		// `Subscribe Now` is selected on marketplace UI
		ImageId:      aws.String("ami-022667efd26192f0b"), // openSUSE Leap 15.2
		InstanceType: types.InstanceTypeT3Nano,
		MinCount:     aws.Int32(1),
		MaxCount:     aws.Int32(1),
		TagSpecifications: []types.TagSpecification{
			{
				ResourceType: types.ResourceTypeInstance,
				// TODO: Create a convenience function to add shared tags to the resources
				Tags: resourceTags(instanceSrc, testID),
			},
		},
	}

	result, err := client.RunInstances(context.Background(), input)
	if err != nil {
		return err
	}

	waiter := ec2.NewInstanceRunningWaiter(client)
	err = waiter.Wait(context.Background(), &ec2.DescribeInstancesInput{
		InstanceIds: []string{*result.Instances[0].InstanceId},
	},
		5*time.Minute)
	if err != nil {
		return err
	}

	return nil
}

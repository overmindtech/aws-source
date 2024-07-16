package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/overmindtech/aws-source/sources/integration"
)

// findActiveInstanceIDByTags finds an instance by tags
// additionalAttr is a variadic parameter that allows to specify additional attributes to search for
// it ignores terminated instances
func findActiveInstanceIDByTags(client *ec2.Client, additionalAttr ...string) (*string, error) {
	result, err := client.DescribeInstances(context.Background(), &ec2.DescribeInstancesInput{})
	if err != nil {
		return nil, err
	}

	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			// ignore terminated or shutting down instances
			if instance.State.Name == types.InstanceStateNameTerminated ||
				instance.State.Name == types.InstanceStateNameShuttingDown {
				// ignore terminated instances
				continue
			}

			if hasTags(instance.Tags, resourceTags(instanceSrc, integration.TestID(), additionalAttr...)) {
				return instance.InstanceId, nil
			}
		}
	}

	return nil, integration.NewNotFoundError(integration.ResourceName(integration.EC2, instanceSrc, additionalAttr...))
}

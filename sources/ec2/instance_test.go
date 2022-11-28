package ec2

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/discovery"
)

type TestResources struct {
	InstanceID string
}

// createEC2 Creates the EC2 resource required for testing
func createEC2(t *testing.T) TestResources {
	var err error
	ec2Client := ec2.NewFromConfig(TestAWSConfig)

	filterName := "name"

	// Find the image ID
	imagesOutput, err := ec2Client.DescribeImages(
		context.Background(),
		&ec2.DescribeImagesInput{
			Filters: []types.Filter{
				{
					Name: &filterName,
					Values: []string{
						"amzn2-ami-kernel-*-x86_64-gp2",
					},
				},
			},
			Owners: []string{
				"amazon",
			},
		},
	)

	if err != nil {
		t.Fatal(err)
	}

	images := imagesOutput.Images

	sort.Slice(
		images,
		func(i, j int) bool {
			iCreation, _ := time.Parse("2006-01-02T15:04:05.000Z", *images[i].CreationDate)
			jCreation, _ := time.Parse("2006-01-02T15:04:05.000Z", *images[j].CreationDate)
			return iCreation.After(jCreation)
		},
	)

	// Get the most recent image
	image := images[0]
	var count int32 = 1
	var runInstancesOutput *ec2.RunInstancesOutput

	runInstancesOutput, err = ec2Client.RunInstances(
		context.Background(),
		&ec2.RunInstancesInput{
			MaxCount:     &count,
			MinCount:     &count,
			ImageId:      image.ImageId,
			InstanceType: types.InstanceTypeT3Micro,
			SubnetId:     TestVPC.Subnets[0].ID,
			TagSpecifications: []types.TagSpecification{
				{
					ResourceType: types.ResourceTypeInstance,
					Tags:         sources.TestTags,
				},
			},
		},
	)

	if err != nil {
		t.Fatal(err)
	}

	instanceID := runInstancesOutput.Instances[0].InstanceId

	t.Cleanup(func() {
		_, err := ec2Client.TerminateInstances(
			context.Background(),
			&ec2.TerminateInstancesInput{
				InstanceIds: []string{
					*instanceID,
				},
			},
		)

		if err != nil {
			t.Error(err)
		}
	})

	return TestResources{
		InstanceID: *instanceID,
	}
}

func TestEC2(t *testing.T) {
	t.Parallel()

	tr := createEC2(t)

	src := InstanceSource{
		Config:    TestAWSConfig,
		AccountID: TestAccountID,
	}

	t.Run("Get with correct instance ID", func(t *testing.T) {
		item, err := src.Get(context.Background(), TestContext, tr.InstanceID)

		if err != nil {
			t.Fatal(err)
		}

		discovery.TestValidateItem(t, item)
	})

	t.Run("Get with incorrect instance ID", func(t *testing.T) {
		_, err := src.Get(context.Background(), TestContext, "i-0ecfa0a234cbc132")

		if err == nil {
			t.Error("expected error but got nil")
		}
	})

	t.Run("Find", func(t *testing.T) {
		items, err := src.Find(context.Background(), TestContext)

		if err != nil {
			t.Error(err)
		}

		if len(items) == 0 {
			t.Error("Expected items to be found but got nothing")
		}

		discovery.TestValidateItems(t, items)
	})

}

package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

// ImageInputMapperGet Gets a given image. As opposed to list, get will get
// details of any image given a correct ID, not just images owned by the current
// account
func imageInputMapperGet(scope string, query string) (*ec2.DescribeImagesInput, error) {
	return &ec2.DescribeImagesInput{
		ImageIds: []string{
			query,
		},
	}, nil
}

// ImageInputMapperList Lists images that are owned by the current account, as
// opposed to all available images since this is simply way too much data
func imageInputMapperList(scope string) (*ec2.DescribeImagesInput, error) {
	return &ec2.DescribeImagesInput{
		Owners: []string{
			// Avoid getting every image in existence, just get the ones
			// relevant to this scope i.e. owned by this account in this region
			"self",
		},
	}, nil
}

func imageOutputMapper(_ context.Context, _ *ec2.Client, scope string, _ *ec2.DescribeImagesInput, output *ec2.DescribeImagesOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, image := range output.Images {
		var err error
		var attrs *sdp.ItemAttributes
		attrs, err = sources.ToAttributesCase(image)

		if err != nil {
			return nil, &sdp.QueryError{
				ErrorType:   sdp.QueryError_OTHER,
				ErrorString: err.Error(),
				Scope:       scope,
			}
		}

		item := sdp.Item{
			Type:            "ec2-image",
			UniqueAttribute: "imageId",
			Scope:           scope,
			Attributes:      attrs,
			Tags:            tagsToMap(image.Tags),
		}

		items = append(items, &item)
	}

	return items, nil
}

//go:generate docgen ../../docs-data
// +overmind:type ec2-image
// +overmind:descriptiveType Amazon Machine Image (AMI)
// +overmind:get Get an AMI by ID
// +overmind:list List all AMIs
// +overmind:search Search AMIs by ARN
// +overmind:group AWS
// +overmind:terraform:queryMap aws_ami.id

func NewImageSource(config aws.Config, accountID string, limit *sources.LimitBucket) *sources.DescribeOnlySource[*ec2.DescribeImagesInput, *ec2.DescribeImagesOutput, *ec2.Client, *ec2.Options] {
	return &sources.DescribeOnlySource[*ec2.DescribeImagesInput, *ec2.DescribeImagesOutput, *ec2.Client, *ec2.Options]{
		Config:    config,
		Client:    ec2.NewFromConfig(config),
		AccountID: accountID,
		ItemType:  "ec2-image",
		DescribeFunc: func(ctx context.Context, client *ec2.Client, input *ec2.DescribeImagesInput) (*ec2.DescribeImagesOutput, error) {
			limit.Wait(ctx) // Wait for rate limiting // Wait for late limiting
			return client.DescribeImages(ctx, input)
		},
		InputMapperGet:  imageInputMapperGet,
		InputMapperList: imageInputMapperList,
		OutputMapper:    imageOutputMapper,
	}
}

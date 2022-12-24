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
func ImageInputMapperGet(scope string, query string) (*ec2.DescribeImagesInput, error) {
	return &ec2.DescribeImagesInput{
		ImageIds: []string{
			query,
		},
	}, nil
}

// ImageInputMapperList Lists images that are owned by the current account, as
// opposed to all available images since this is simply way too much data
func ImageInputMapperList(scope string) (*ec2.DescribeImagesInput, error) {
	return &ec2.DescribeImagesInput{
		Owners: []string{
			// Avoid getting every image in existence, just get the ones
			// relevant to this scope i.e. owned by this account in this region
			"self",
		},
	}, nil
}

func ImageOutputMapper(scope string, output *ec2.DescribeImagesOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, Image := range output.Images {
		var err error
		var attrs *sdp.ItemAttributes
		attrs, err = sources.ToAttributesCase(Image)

		if err != nil {
			return nil, &sdp.ItemRequestError{
				ErrorType:   sdp.ItemRequestError_OTHER,
				ErrorString: err.Error(),
				Scope:       scope,
			}
		}

		item := sdp.Item{
			Type:            "ec2-image",
			UniqueAttribute: "imageId",
			Scope:           scope,
			Attributes:      attrs,
		}

		items = append(items, &item)
	}

	return items, nil
}

func NewImageSource(config aws.Config, accountID string) *sources.AWSSource[*ec2.DescribeImagesInput, *ec2.DescribeImagesOutput, *ec2.Client, *ec2.Options] {
	return &sources.AWSSource[*ec2.DescribeImagesInput, *ec2.DescribeImagesOutput, *ec2.Client, *ec2.Options]{
		Config:    config,
		AccountID: accountID,
		ItemType:  "ec2-image",
		DescribeFunc: func(ctx context.Context, client *ec2.Client, input *ec2.DescribeImagesInput) (*ec2.DescribeImagesOutput, error) {
			return client.DescribeImages(ctx, input)
		},
		InputMapperGet:  ImageInputMapperGet,
		InputMapperList: ImageInputMapperList,
		OutputMapper:    ImageOutputMapper,
	}
}

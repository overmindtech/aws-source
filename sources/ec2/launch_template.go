package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func LaunchTemplateInputMapperGet(scope string, query string) (*ec2.DescribeLaunchTemplatesInput, error) {
	return &ec2.DescribeLaunchTemplatesInput{
		LaunchTemplateIds: []string{
			query,
		},
	}, nil
}

func LaunchTemplateInputMapperList(scope string) (*ec2.DescribeLaunchTemplatesInput, error) {
	return &ec2.DescribeLaunchTemplatesInput{}, nil
}

func LaunchTemplateOutputMapper(scope string, output *ec2.DescribeLaunchTemplatesOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, LaunchTemplate := range output.LaunchTemplates {
		var err error
		var attrs *sdp.ItemAttributes
		attrs, err = sources.ToAttributesCase(LaunchTemplate)

		if err != nil {
			return nil, &sdp.ItemRequestError{
				ErrorType:   sdp.ItemRequestError_OTHER,
				ErrorString: err.Error(),
				Scope:       scope,
			}
		}

		item := sdp.Item{
			Type:            "ec2-launch-template",
			UniqueAttribute: "launchTemplateId",
			Scope:           scope,
			Attributes:      attrs,
		}

		items = append(items, &item)
	}

	return items, nil
}

func NewLaunchTemplateSource(config aws.Config, accountID string) *EC2Source[*ec2.DescribeLaunchTemplatesInput, *ec2.DescribeLaunchTemplatesOutput] {
	return &EC2Source[*ec2.DescribeLaunchTemplatesInput, *ec2.DescribeLaunchTemplatesOutput]{
		Config:    config,
		AccountID: accountID,
		ItemType:  "ec2-launch-template",
		DescribeFunc: func(ctx context.Context, client *ec2.Client, input *ec2.DescribeLaunchTemplatesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeLaunchTemplatesOutput, error) {
			return client.DescribeLaunchTemplates(ctx, input)
		},
		InputMapperGet:  LaunchTemplateInputMapperGet,
		InputMapperList: LaunchTemplateInputMapperList,
		PaginatorBuilder: func(client *ec2.Client, params *ec2.DescribeLaunchTemplatesInput) Paginator[*ec2.DescribeLaunchTemplatesOutput] {
			return ec2.NewDescribeLaunchTemplatesPaginator(client, params)
		},
		OutputMapper: LaunchTemplateOutputMapper,
	}
}

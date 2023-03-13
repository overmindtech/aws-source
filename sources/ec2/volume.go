package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func VolumeInputMapperGet(scope string, query string) (*ec2.DescribeVolumesInput, error) {
	return &ec2.DescribeVolumesInput{
		VolumeIds: []string{
			query,
		},
	}, nil
}

func VolumeInputMapperList(scope string) (*ec2.DescribeVolumesInput, error) {
	return &ec2.DescribeVolumesInput{}, nil
}

func VolumeOutputMapper(scope string, _ *ec2.DescribeVolumesInput, output *ec2.DescribeVolumesOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, volume := range output.Volumes {
		var err error
		var attrs *sdp.ItemAttributes
		attrs, err = sources.ToAttributesCase(volume)

		if err != nil {
			return nil, &sdp.QueryError{
				ErrorType:   sdp.QueryError_OTHER,
				ErrorString: err.Error(),
				Scope:       scope,
			}
		}

		item := sdp.Item{
			Type:            "ec2-volume",
			UniqueAttribute: "volumeId",
			Scope:           scope,
			Attributes:      attrs,
		}

		for _, attachment := range volume.Attachments {
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
				Type:   "ec2-instance",
				Method: sdp.QueryMethod_GET,
				Query:  *attachment.InstanceId,
				Scope:  scope,
			})
		}

		items = append(items, &item)
	}

	return items, nil
}

func NewVolumeSource(config aws.Config, accountID string, limit *LimitBucket) *sources.DescribeOnlySource[*ec2.DescribeVolumesInput, *ec2.DescribeVolumesOutput, *ec2.Client, *ec2.Options] {
	return &sources.DescribeOnlySource[*ec2.DescribeVolumesInput, *ec2.DescribeVolumesOutput, *ec2.Client, *ec2.Options]{
		Config:    config,
		Client:    ec2.NewFromConfig(config),
		AccountID: accountID,
		ItemType:  "ec2-volume",
		DescribeFunc: func(ctx context.Context, client *ec2.Client, input *ec2.DescribeVolumesInput) (*ec2.DescribeVolumesOutput, error) {
			<-limit.C // Wait for late limiting
			return client.DescribeVolumes(ctx, input)
		},
		InputMapperGet:  VolumeInputMapperGet,
		InputMapperList: VolumeInputMapperList,
		PaginatorBuilder: func(client *ec2.Client, params *ec2.DescribeVolumesInput) sources.Paginator[*ec2.DescribeVolumesOutput, *ec2.Options] {
			return ec2.NewDescribeVolumesPaginator(client, params)
		},
		OutputMapper: VolumeOutputMapper,
	}
}

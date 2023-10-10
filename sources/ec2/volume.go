package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func volumeInputMapperGet(scope string, query string) (*ec2.DescribeVolumesInput, error) {
	return &ec2.DescribeVolumesInput{
		VolumeIds: []string{
			query,
		},
	}, nil
}

func volumeInputMapperList(scope string) (*ec2.DescribeVolumesInput, error) {
	return &ec2.DescribeVolumesInput{}, nil
}

func volumeOutputMapper(_ context.Context, _ *ec2.Client, scope string, _ *ec2.DescribeVolumesInput, output *ec2.DescribeVolumesOutput) ([]*sdp.Item, error) {
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
			Tags:            tagsToMap(volume.Tags),
		}

		for _, attachment := range volume.Attachments {
			// +overmind:link ec2-instance
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "ec2-instance",
					Method: sdp.QueryMethod_GET,
					Query:  *attachment.InstanceId,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					// The instance and the volume are closely linked
					In:  true,
					Out: true,
				},
			})
		}

		if volume.AvailabilityZone != nil {
			// +overmind:link ec2-availability-zone
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "ec2-availability-zone",
					Method: sdp.QueryMethod_GET,
					Query:  *volume.AvailabilityZone,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					// AZs don't change
					In:  false,
					Out: false,
				},
			})
		}

		items = append(items, &item)
	}

	return items, nil
}

//go:generate docgen ../../docs-data
// +overmind:type ec2-volume
// +overmind:descriptiveType EC2 Volume
// +overmind:get Get a volume by ID
// +overmind:list List all volumes
// +overmind:search Search volumes by ARN
// +overmind:group AWS
// +overmind:terraform:queryMap aws_ebs_volume.id

func NewVolumeSource(config aws.Config, accountID string, limit *sources.LimitBucket) *sources.DescribeOnlySource[*ec2.DescribeVolumesInput, *ec2.DescribeVolumesOutput, *ec2.Client, *ec2.Options] {
	return &sources.DescribeOnlySource[*ec2.DescribeVolumesInput, *ec2.DescribeVolumesOutput, *ec2.Client, *ec2.Options]{
		Config:    config,
		Client:    ec2.NewFromConfig(config),
		AccountID: accountID,
		ItemType:  "ec2-volume",
		DescribeFunc: func(ctx context.Context, client *ec2.Client, input *ec2.DescribeVolumesInput) (*ec2.DescribeVolumesOutput, error) {
			limit.Wait(ctx) // Wait for rate limiting // Wait for late limiting
			return client.DescribeVolumes(ctx, input)
		},
		InputMapperGet:  volumeInputMapperGet,
		InputMapperList: volumeInputMapperList,
		PaginatorBuilder: func(client *ec2.Client, params *ec2.DescribeVolumesInput) sources.Paginator[*ec2.DescribeVolumesOutput, *ec2.Options] {
			return ec2.NewDescribeVolumesPaginator(client, params)
		},
		OutputMapper: volumeOutputMapper,
	}
}

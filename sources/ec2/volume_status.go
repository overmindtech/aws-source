package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func volumeStatusInputMapperGet(scope string, query string) (*ec2.DescribeVolumeStatusInput, error) {
	return &ec2.DescribeVolumeStatusInput{
		VolumeIds: []string{
			query,
		},
	}, nil
}

func volumeStatusInputMapperList(scope string) (*ec2.DescribeVolumeStatusInput, error) {
	return &ec2.DescribeVolumeStatusInput{}, nil
}

func volumeStatusOutputMapper(scope string, _ *ec2.DescribeVolumeStatusInput, output *ec2.DescribeVolumeStatusOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, volume := range output.VolumeStatuses {
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
			Type:            "ec2-volume-status",
			UniqueAttribute: "volumeId",
			Scope:           scope,
			Attributes:      attrs,
			LinkedItemQueries: []*sdp.LinkedItemQuery{
				{
					Query: &sdp.Query{
						// Always get the volume
						Type:   "ec2-volume",
						Method: sdp.QueryMethod_GET,
						Query:  *volume.VolumeId,
						Scope:  scope,
					},
				},
			},
		}

		if volume.VolumeStatus != nil {
			switch volume.VolumeStatus.Status {
			case types.VolumeStatusInfoStatusImpaired:
				item.Health = sdp.Health_HEALTH_ERROR.Enum()
			case types.VolumeStatusInfoStatusOk:
				item.Health = sdp.Health_HEALTH_OK.Enum()
			case types.VolumeStatusInfoStatusInsufficientData:
				item.Health = sdp.Health_HEALTH_UNKNOWN.Enum()
			}
		}

		for _, event := range volume.Events {
			if event.InstanceId != nil {
				// +overmind:link ec2-instance
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{Query: &sdp.Query{
					Type:   "ec2-instance",
					Method: sdp.QueryMethod_GET,
					Query:  *event.InstanceId,
					Scope:  scope,
				}})
			}
		}

		items = append(items, &item)
	}

	return items, nil
}

//go:generate docgen ../../docs-data
// +overmind:type ec2-volume-status
// +overmind:descriptiveType EC2 Volume Status
// +overmind:get Get a volume status by volume ID
// +overmind:list List all volume statuses
// +overmind:search Search for volume statuses by ARN
// +overmind:group AWS

func NewVolumeStatusSource(config aws.Config, accountID string, limit *sources.LimitBucket) *sources.DescribeOnlySource[*ec2.DescribeVolumeStatusInput, *ec2.DescribeVolumeStatusOutput, *ec2.Client, *ec2.Options] {
	return &sources.DescribeOnlySource[*ec2.DescribeVolumeStatusInput, *ec2.DescribeVolumeStatusOutput, *ec2.Client, *ec2.Options]{
		Config:    config,
		Client:    ec2.NewFromConfig(config),
		AccountID: accountID,
		ItemType:  "ec2-volume-status",
		DescribeFunc: func(ctx context.Context, client *ec2.Client, input *ec2.DescribeVolumeStatusInput) (*ec2.DescribeVolumeStatusOutput, error) {
			<-limit.C // Wait for late limiting
			return client.DescribeVolumeStatus(ctx, input)
		},
		InputMapperGet:  volumeStatusInputMapperGet,
		InputMapperList: volumeStatusInputMapperList,
		PaginatorBuilder: func(client *ec2.Client, params *ec2.DescribeVolumeStatusInput) sources.Paginator[*ec2.DescribeVolumeStatusOutput, *ec2.Options] {
			return ec2.NewDescribeVolumeStatusPaginator(client, params)
		},
		OutputMapper: volumeStatusOutputMapper,
	}
}

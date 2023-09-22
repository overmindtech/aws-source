package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func snapshotInputMapperGet(scope string, query string) (*ec2.DescribeSnapshotsInput, error) {
	return &ec2.DescribeSnapshotsInput{
		SnapshotIds: []string{
			query,
		},
	}, nil
}

func snapshotInputMapperList(scope string) (*ec2.DescribeSnapshotsInput, error) {
	return &ec2.DescribeSnapshotsInput{
		OwnerIds: []string{
			// Avoid getting every snapshot in existence, just get the ones
			// relevant to this scope i.e. owned by this account in this region
			"self",
		},
	}, nil
}

func snapshotOutputMapper(scope string, _ *ec2.DescribeSnapshotsInput, output *ec2.DescribeSnapshotsOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, snapshot := range output.Snapshots {
		var err error
		var attrs *sdp.ItemAttributes
		attrs, err = sources.ToAttributesCase(snapshot)

		if err != nil {
			return nil, &sdp.QueryError{
				ErrorType:   sdp.QueryError_OTHER,
				ErrorString: err.Error(),
				Scope:       scope,
			}
		}

		item := sdp.Item{
			Type:            "ec2-snapshot",
			UniqueAttribute: "snapshotId",
			Scope:           scope,
			Attributes:      attrs,
		}

		if snapshot.VolumeId != nil {
			// Ignore the arbitrary ID that is used by Amazon
			if *snapshot.VolumeId != "vol-ffffffff" {
				// +overmind:link ec2-volume
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "ec2-volume",
						Method: sdp.QueryMethod_GET,
						Query:  *snapshot.VolumeId,
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						// Changing the volume will probably affect the snapshot
						In: true,
						// Changing the snapshot will affect the volume indirectly
						// as applications might rely on snapshots as backups
						// or other use-cases
						Out: true,
					},
				})
			}
		}

		items = append(items, &item)
	}

	return items, nil
}

//go:generate docgen ../../docs-data
// +overmind:type ec2-snapshot
// +overmind:descriptiveType EC2 Snapshot
// +overmind:get Get a snapshot by ID
// +overmind:list List all snapshots
// +overmind:search Search snapshots by ARN
// +overmind:group AWS

func NewSnapshotSource(config aws.Config, accountID string, limit *sources.LimitBucket) *sources.DescribeOnlySource[*ec2.DescribeSnapshotsInput, *ec2.DescribeSnapshotsOutput, *ec2.Client, *ec2.Options] {
	return &sources.DescribeOnlySource[*ec2.DescribeSnapshotsInput, *ec2.DescribeSnapshotsOutput, *ec2.Client, *ec2.Options]{
		Config:    config,
		Client:    ec2.NewFromConfig(config),
		AccountID: accountID,
		ItemType:  "ec2-snapshot",
		DescribeFunc: func(ctx context.Context, client *ec2.Client, input *ec2.DescribeSnapshotsInput) (*ec2.DescribeSnapshotsOutput, error) {
			limit.Wait(ctx) // Wait for rate limiting // Wait for late limiting
			return client.DescribeSnapshots(ctx, input)
		},
		InputMapperGet:  snapshotInputMapperGet,
		InputMapperList: snapshotInputMapperList,
		PaginatorBuilder: func(client *ec2.Client, params *ec2.DescribeSnapshotsInput) sources.Paginator[*ec2.DescribeSnapshotsOutput, *ec2.Options] {
			return ec2.NewDescribeSnapshotsPaginator(client, params)
		},
		OutputMapper: snapshotOutputMapper,
	}
}

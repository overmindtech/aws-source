package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func placementGroupInputMapperGet(scope string, query string) (*ec2.DescribePlacementGroupsInput, error) {
	return &ec2.DescribePlacementGroupsInput{
		GroupIds: []string{
			query,
		},
	}, nil
}

func placementGroupInputMapperList(scope string) (*ec2.DescribePlacementGroupsInput, error) {
	return &ec2.DescribePlacementGroupsInput{}, nil
}

func placementGroupOutputMapper(_ context.Context, _ *ec2.Client, scope string, _ *ec2.DescribePlacementGroupsInput, output *ec2.DescribePlacementGroupsOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, ng := range output.PlacementGroups {
		var err error
		var attrs *sdp.ItemAttributes
		attrs, err = sources.ToAttributesCase(ng, "tags")

		if err != nil {
			return nil, &sdp.QueryError{
				ErrorType:   sdp.QueryError_OTHER,
				ErrorString: err.Error(),
				Scope:       scope,
			}
		}

		item := sdp.Item{
			Type:            "ec2-placement-group",
			UniqueAttribute: "groupId",
			Scope:           scope,
			Attributes:      attrs,
			Tags:            tagsToMap(ng.Tags),
		}

		items = append(items, &item)
	}

	return items, nil
}

//go:generate docgen ../../docs-data
// +overmind:type ec2-placement-group
// +overmind:descriptiveType Placement Group
// +overmind:get Get a placement group by ID
// +overmind:list List all placement groups
// +overmind:search Search for placement groups by ARN
// +overmind:group AWS
// +overmind:terraform:queryMap aws_placement_group.id

func NewPlacementGroupSource(client *ec2.Client, accountID string, region string, limit *sources.LimitBucket) *sources.DescribeOnlySource[*ec2.DescribePlacementGroupsInput, *ec2.DescribePlacementGroupsOutput, *ec2.Client, *ec2.Options] {
	return &sources.DescribeOnlySource[*ec2.DescribePlacementGroupsInput, *ec2.DescribePlacementGroupsOutput, *ec2.Client, *ec2.Options]{

		Client:    client,
		AccountID: accountID,
		ItemType:  "ec2-placement-group",
		DescribeFunc: func(ctx context.Context, client *ec2.Client, input *ec2.DescribePlacementGroupsInput) (*ec2.DescribePlacementGroupsOutput, error) {
			limit.Wait(ctx) // Wait for rate limiting // Wait for late limiting
			return client.DescribePlacementGroups(ctx, input)
		},
		InputMapperGet:  placementGroupInputMapperGet,
		InputMapperList: placementGroupInputMapperList,
		OutputMapper:    placementGroupOutputMapper,
	}
}

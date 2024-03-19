package rds

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func dBSubnetGroupOutputMapper(ctx context.Context, client rdsClient, scope string, _ *rds.DescribeDBSubnetGroupsInput, output *rds.DescribeDBSubnetGroupsOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, sg := range output.DBSubnetGroups {
		var tags map[string]string

		// Get tags
		tagsOut, err := client.ListTagsForResource(ctx, &rds.ListTagsForResourceInput{
			ResourceName: sg.DBSubnetGroupArn,
		})

		if err == nil {
			tags = tagsToMap(tagsOut.TagList)
		} else {
			tags = sources.HandleTagsError(ctx, err)
		}

		attributes, err := sources.ToAttributesCase(sg)

		if err != nil {
			return nil, err
		}

		item := sdp.Item{
			Type:            "rds-db-subnet-group",
			UniqueAttribute: "dBSubnetGroupName",
			Attributes:      attributes,
			Scope:           scope,
			Tags:            tags,
		}

		var a *sources.ARN

		if sg.VpcId != nil {
			// +overmind:link ec2-vpc
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "ec2-vpc",
					Method: sdp.QueryMethod_GET,
					Query:  *sg.VpcId,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					// Changing the VPC can affect the subnet group
					In: true,
					// The subnet group won't affect the VPC
					Out: false,
				},
			})
		}

		for _, subnet := range sg.Subnets {
			if subnet.SubnetIdentifier != nil {
				// +overmind:link ec2-subnet
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "ec2-subnet",
						Method: sdp.QueryMethod_GET,
						Query:  *subnet.SubnetIdentifier,
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						// Changing the subnet can affect the subnet group
						In: true,
						// The subnet group won't affect the subnet
						Out: false,
					},
				})
			}

			if subnet.SubnetOutpost != nil {
				if subnet.SubnetOutpost.Arn != nil {
					if a, err = sources.ParseARN(*subnet.SubnetOutpost.Arn); err == nil {
						// +overmind:link outposts-outpost
						item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
							Query: &sdp.Query{
								Type:   "outposts-outpost",
								Method: sdp.QueryMethod_SEARCH,
								Query:  *subnet.SubnetOutpost.Arn,
								Scope:  sources.FormatScope(a.AccountID, a.Region),
							},
							BlastPropagation: &sdp.BlastPropagation{
								// Changing the outpost can affect the subnet group
								In: true,
								// The subnet group won't affect the outpost
								Out: false,
							},
						})
					}
				}
			}
		}

		items = append(items, &item)
	}

	return items, nil
}

//go:generate docgen ../../docs-data
// +overmind:type rds-db-subnet-group
// +overmind:descriptiveType RDS Subnet Group
// +overmind:get Get a subnet group by name
// +overmind:list List all subnet groups
// +overmind:search Search for subnet groups by ARN
// +overmind:group AWS
// +overmind:terraform:queryMap aws_db_subnet_group.arn
// +overmind:terraform:method SEARCH

func NewDBSubnetGroupSource(client rdsClient, accountID string, region string) *sources.DescribeOnlySource[*rds.DescribeDBSubnetGroupsInput, *rds.DescribeDBSubnetGroupsOutput, rdsClient, *rds.Options] {
	return &sources.DescribeOnlySource[*rds.DescribeDBSubnetGroupsInput, *rds.DescribeDBSubnetGroupsOutput, rdsClient, *rds.Options]{
		ItemType: "rds-db-subnet-group",

		AccountID: accountID,
		Client:    client,
		PaginatorBuilder: func(client rdsClient, params *rds.DescribeDBSubnetGroupsInput) sources.Paginator[*rds.DescribeDBSubnetGroupsOutput, *rds.Options] {
			return rds.NewDescribeDBSubnetGroupsPaginator(client, params)
		},
		DescribeFunc: func(ctx context.Context, client rdsClient, input *rds.DescribeDBSubnetGroupsInput) (*rds.DescribeDBSubnetGroupsOutput, error) {
			return client.DescribeDBSubnetGroups(ctx, input)
		},
		InputMapperGet: func(scope, query string) (*rds.DescribeDBSubnetGroupsInput, error) {
			return &rds.DescribeDBSubnetGroupsInput{
				DBSubnetGroupName: &query,
			}, nil
		},
		InputMapperList: func(scope string) (*rds.DescribeDBSubnetGroupsInput, error) {
			return &rds.DescribeDBSubnetGroupsInput{}, nil
		},
		OutputMapper: dBSubnetGroupOutputMapper,
	}
}

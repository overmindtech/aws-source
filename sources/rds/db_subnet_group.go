package rds

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func dBSubnetGroupOutputMapper(scope string, _ *rds.DescribeDBSubnetGroupsInput, output *rds.DescribeDBSubnetGroupsOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, sg := range output.DBSubnetGroups {
		attributes, err := sources.ToAttributesCase(sg)

		if err != nil {
			return nil, err
		}

		item := sdp.Item{
			Type:            "rds-db-subnet-group",
			UniqueAttribute: "dBSubnetGroupName",
			Attributes:      attributes,
			Scope:           scope,
		}

		var a *sources.ARN

		if sg.VpcId != nil {
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
				Type:   "ec2-vpc",
				Method: sdp.QueryMethod_GET,
				Query:  *sg.VpcId,
				Scope:  scope,
			})
		}

		for _, subnet := range sg.Subnets {
			if subnet.SubnetIdentifier != nil {
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
					Type:   "ec2-subnet",
					Method: sdp.QueryMethod_GET,
					Query:  *subnet.SubnetIdentifier,
					Scope:  scope,
				})
			}

			if subnet.SubnetAvailabilityZone != nil {
				if subnet.SubnetAvailabilityZone.Name != nil {
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
						Type:   "ec2-availability-zone",
						Method: sdp.QueryMethod_GET,
						Query:  *subnet.SubnetAvailabilityZone.Name,
						Scope:  scope,
					})
				}
			}

			if subnet.SubnetOutpost != nil {
				if subnet.SubnetOutpost.Arn != nil {
					if a, err = sources.ParseARN(*subnet.SubnetOutpost.Arn); err == nil {
						item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
							Type:   "outposts-outpost",
							Method: sdp.QueryMethod_SEARCH,
							Query:  *subnet.SubnetOutpost.Arn,
							Scope:  sources.FormatScope(a.AccountID, a.Region),
						})
					}
				}
			}
		}

		items = append(items, &item)
	}

	return items, nil
}

func NewDBSubnetGroupSource(config aws.Config, accountID string) *sources.DescribeOnlySource[*rds.DescribeDBSubnetGroupsInput, *rds.DescribeDBSubnetGroupsOutput, *rds.Client, *rds.Options] {
	return &sources.DescribeOnlySource[*rds.DescribeDBSubnetGroupsInput, *rds.DescribeDBSubnetGroupsOutput, *rds.Client, *rds.Options]{
		ItemType:  "rds-db-subnet-group",
		Config:    config,
		AccountID: accountID,
		Client:    rds.NewFromConfig(config),
		PaginatorBuilder: func(client *rds.Client, params *rds.DescribeDBSubnetGroupsInput) sources.Paginator[*rds.DescribeDBSubnetGroupsOutput, *rds.Options] {
			return rds.NewDescribeDBSubnetGroupsPaginator(client, params)
		},
		DescribeFunc: func(ctx context.Context, client *rds.Client, input *rds.DescribeDBSubnetGroupsInput) (*rds.DescribeDBSubnetGroupsOutput, error) {
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

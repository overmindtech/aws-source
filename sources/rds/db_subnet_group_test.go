package rds

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func TestDBSubnetGroupOutputMapper(t *testing.T) {
	output := rds.DescribeDBSubnetGroupsOutput{
		DBSubnetGroups: []types.DBSubnetGroup{
			{
				DBSubnetGroupName:        sources.PtrString("default-vpc-0d7892e00e573e701"),
				DBSubnetGroupDescription: sources.PtrString("Created from the RDS Management Console"),
				VpcId:                    sources.PtrString("vpc-0d7892e00e573e701"), // link
				SubnetGroupStatus:        sources.PtrString("Complete"),
				Subnets: []types.Subnet{
					{
						SubnetIdentifier: sources.PtrString("subnet-0450a637af9984235"), // link
						SubnetAvailabilityZone: &types.AvailabilityZone{
							Name: sources.PtrString("eu-west-2c"), // link
						},
						SubnetOutpost: &types.Outpost{
							Arn: sources.PtrString("arn:aws:service:region:account:type/id"), // link
						},
						SubnetStatus: sources.PtrString("Active"),
					},
				},
				DBSubnetGroupArn: sources.PtrString("arn:aws:rds:eu-west-2:052392120703:subgrp:default-vpc-0d7892e00e573e701"),
				SupportedNetworkTypes: []string{
					"IPV4",
				},
			},
		},
	}

	items, err := dBSubnetGroupOutputMapper(context.Background(), mockRdsClient{}, "foo", nil, &output)

	if err != nil {
		t.Fatal(err)
	}

	if len(items) != 1 {
		t.Fatalf("got %v items, expected 1", len(items))
	}

	item := items[0]

	if err = item.Validate(); err != nil {
		t.Error(err)
	}

	if item.Tags["key"] != "value" {
		t.Errorf("expected key to be value, got %v", item.Tags["key"])
	}

	tests := sources.QueryTests{
		{
			ExpectedType:   "ec2-vpc",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "vpc-0d7892e00e573e701",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "ec2-subnet",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "subnet-0450a637af9984235",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "outposts-outpost",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "arn:aws:service:region:account:type/id",
			ExpectedScope:  "account.region",
		},
	}

	tests.Execute(t, item)
}

func TestNewDBSubnetGroupSource(t *testing.T) {
	client, account, region := GetAutoConfig(t)

	source := NewDBSubnetGroupSource(client, account, region)

	test := sources.E2ETest{
		Source:  source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}

package rds

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/overmindtech/aws-source/adapters"
	"github.com/overmindtech/sdp-go"
)

func TestDBSubnetGroupOutputMapper(t *testing.T) {
	output := rds.DescribeDBSubnetGroupsOutput{
		DBSubnetGroups: []types.DBSubnetGroup{
			{
				DBSubnetGroupName:        adapters.PtrString("default-vpc-0d7892e00e573e701"),
				DBSubnetGroupDescription: adapters.PtrString("Created from the RDS Management Console"),
				VpcId:                    adapters.PtrString("vpc-0d7892e00e573e701"), // link
				SubnetGroupStatus:        adapters.PtrString("Complete"),
				Subnets: []types.Subnet{
					{
						SubnetIdentifier: adapters.PtrString("subnet-0450a637af9984235"), // link
						SubnetAvailabilityZone: &types.AvailabilityZone{
							Name: adapters.PtrString("eu-west-2c"), // link
						},
						SubnetOutpost: &types.Outpost{
							Arn: adapters.PtrString("arn:aws:service:region:account:type/id"), // link
						},
						SubnetStatus: adapters.PtrString("Active"),
					},
				},
				DBSubnetGroupArn: adapters.PtrString("arn:aws:rds:eu-west-2:052392120703:subgrp:default-vpc-0d7892e00e573e701"),
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

	if item.GetTags()["key"] != "value" {
		t.Errorf("expected key to be value, got %v", item.GetTags()["key"])
	}

	tests := adapters.QueryTests{
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

	test := adapters.E2ETest{
		Adapter: source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}

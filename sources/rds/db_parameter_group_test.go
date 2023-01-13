package rds

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/overmindtech/aws-source/sources"
)

func TestDBParameterGroupOutputMapper(t *testing.T) {
	output := rds.DescribeDBParameterGroupsOutput{
		DBParameterGroups: []types.DBParameterGroup{
			{
				DBParameterGroupName:   sources.PtrString("default.aurora-mysql5.7"),
				DBParameterGroupFamily: sources.PtrString("aurora-mysql5.7"),
				Description:            sources.PtrString("Default parameter group for aurora-mysql5.7"),
				DBParameterGroupArn:    sources.PtrString("arn:aws:rds:eu-west-1:052392120703:pg:default.aurora-mysql5.7"),
			},
		},
	}

	items, err := DBParameterGroupOutputMapper("foo", &output)

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
}

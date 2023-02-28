package rds

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/overmindtech/aws-source/sources"
)

func TestOptionGroupOutputMapper(t *testing.T) {
	output := rds.DescribeOptionGroupsOutput{
		OptionGroupsList: []types.OptionGroup{
			{
				OptionGroupName:                       sources.PtrString("default:aurora-mysql-8-0"),
				OptionGroupDescription:                sources.PtrString("Default option group for aurora-mysql 8.0"),
				EngineName:                            sources.PtrString("aurora-mysql"),
				MajorEngineVersion:                    sources.PtrString("8.0"),
				Options:                               []types.Option{},
				AllowsVpcAndNonVpcInstanceMemberships: true,
				OptionGroupArn:                        sources.PtrString("arn:aws:rds:eu-west-2:052392120703:og:default:aurora-mysql-8-0"),
			},
		},
	}

	items, err := OptionGroupOutputMapper("foo", nil, &output)

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

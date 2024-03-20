package rds

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/overmindtech/aws-source/sources"
)

func TestDBParameterGroupOutputMapper(t *testing.T) {
	group := ParameterGroup{
		DBParameterGroup: types.DBParameterGroup{
			DBParameterGroupName:   sources.PtrString("default.aurora-mysql5.7"),
			DBParameterGroupFamily: sources.PtrString("aurora-mysql5.7"),
			Description:            sources.PtrString("Default parameter group for aurora-mysql5.7"),
			DBParameterGroupArn:    sources.PtrString("arn:aws:rds:eu-west-1:052392120703:pg:default.aurora-mysql5.7"),
		},
		Parameters: []types.Parameter{
			{
				ParameterName:  sources.PtrString("activate_all_roles_on_login"),
				ParameterValue: sources.PtrString("0"),
				Description:    sources.PtrString("Automatically set all granted roles as active after the user has authenticated successfully."),
				Source:         sources.PtrString("engine-default"),
				ApplyType:      sources.PtrString("dynamic"),
				DataType:       sources.PtrString("boolean"),
				AllowedValues:  sources.PtrString("0,1"),
				IsModifiable:   sources.PtrBool(true),
				ApplyMethod:    types.ApplyMethodPendingReboot,
			},
			{
				ParameterName: sources.PtrString("allow-suspicious-udfs"),
				Description:   sources.PtrString("Controls whether user-defined functions that have only an xxx symbol for the main function can be loaded"),
				Source:        sources.PtrString("engine-default"),
				ApplyType:     sources.PtrString("static"),
				DataType:      sources.PtrString("boolean"),
				AllowedValues: sources.PtrString("0,1"),
				IsModifiable:  sources.PtrBool(false),
				ApplyMethod:   types.ApplyMethodPendingReboot,
			},
			{
				ParameterName: sources.PtrString("aurora_parallel_query"),
				Description:   sources.PtrString("This parameter can be used to enable and disable Aurora Parallel Query."),
				Source:        sources.PtrString("engine-default"),
				ApplyType:     sources.PtrString("dynamic"),
				DataType:      sources.PtrString("boolean"),
				AllowedValues: sources.PtrString("0,1"),
				IsModifiable:  sources.PtrBool(true),
				ApplyMethod:   types.ApplyMethodPendingReboot,
			},
			{
				ParameterName: sources.PtrString("autocommit"),
				Description:   sources.PtrString("Sets the autocommit mode"),
				Source:        sources.PtrString("engine-default"),
				ApplyType:     sources.PtrString("dynamic"),
				DataType:      sources.PtrString("boolean"),
				AllowedValues: sources.PtrString("0,1"),
				IsModifiable:  sources.PtrBool(true),
				ApplyMethod:   types.ApplyMethodPendingReboot,
			},
		},
	}

	item, err := dBParameterGroupItemMapper("foo", &group)

	if err != nil {
		t.Fatal(err)
	}

	if err = item.Validate(); err != nil {
		t.Error(err)
	}
}

func TestNewDBParameterGroupSource(t *testing.T) {
	client, account, region := GetAutoConfig(t)

	source := NewDBParameterGroupSource(client, account, region)

	test := sources.E2ETest{
		Source:  source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}

package rds

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/overmindtech/aws-source/adapters"
)

func TestDBParameterGroupOutputMapper(t *testing.T) {
	group := ParameterGroup{
		DBParameterGroup: types.DBParameterGroup{
			DBParameterGroupName:   adapters.PtrString("default.aurora-mysql5.7"),
			DBParameterGroupFamily: adapters.PtrString("aurora-mysql5.7"),
			Description:            adapters.PtrString("Default parameter group for aurora-mysql5.7"),
			DBParameterGroupArn:    adapters.PtrString("arn:aws:rds:eu-west-1:052392120703:pg:default.aurora-mysql5.7"),
		},
		Parameters: []types.Parameter{
			{
				ParameterName:  adapters.PtrString("activate_all_roles_on_login"),
				ParameterValue: adapters.PtrString("0"),
				Description:    adapters.PtrString("Automatically set all granted roles as active after the user has authenticated successfully."),
				Source:         adapters.PtrString("engine-default"),
				ApplyType:      adapters.PtrString("dynamic"),
				DataType:       adapters.PtrString("boolean"),
				AllowedValues:  adapters.PtrString("0,1"),
				IsModifiable:   adapters.PtrBool(true),
				ApplyMethod:    types.ApplyMethodPendingReboot,
			},
			{
				ParameterName: adapters.PtrString("allow-suspicious-udfs"),
				Description:   adapters.PtrString("Controls whether user-defined functions that have only an xxx symbol for the main function can be loaded"),
				Source:        adapters.PtrString("engine-default"),
				ApplyType:     adapters.PtrString("static"),
				DataType:      adapters.PtrString("boolean"),
				AllowedValues: adapters.PtrString("0,1"),
				IsModifiable:  adapters.PtrBool(false),
				ApplyMethod:   types.ApplyMethodPendingReboot,
			},
			{
				ParameterName: adapters.PtrString("aurora_parallel_query"),
				Description:   adapters.PtrString("This parameter can be used to enable and disable Aurora Parallel Query."),
				Source:        adapters.PtrString("engine-default"),
				ApplyType:     adapters.PtrString("dynamic"),
				DataType:      adapters.PtrString("boolean"),
				AllowedValues: adapters.PtrString("0,1"),
				IsModifiable:  adapters.PtrBool(true),
				ApplyMethod:   types.ApplyMethodPendingReboot,
			},
			{
				ParameterName: adapters.PtrString("autocommit"),
				Description:   adapters.PtrString("Sets the autocommit mode"),
				Source:        adapters.PtrString("engine-default"),
				ApplyType:     adapters.PtrString("dynamic"),
				DataType:      adapters.PtrString("boolean"),
				AllowedValues: adapters.PtrString("0,1"),
				IsModifiable:  adapters.PtrBool(true),
				ApplyMethod:   types.ApplyMethodPendingReboot,
			},
		},
	}

	item, err := dBParameterGroupItemMapper("", "foo", &group)

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

	test := adapters.E2ETest{
		Adapter: source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}

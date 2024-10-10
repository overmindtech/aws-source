package rds

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/overmindtech/aws-source/adapters"
)

func TestDBClusterParameterGroupOutputMapper(t *testing.T) {
	group := ClusterParameterGroup{
		DBClusterParameterGroup: types.DBClusterParameterGroup{
			DBClusterParameterGroupName: adapters.PtrString("default.aurora-mysql5.7"),
			DBParameterGroupFamily:      adapters.PtrString("aurora-mysql5.7"),
			Description:                 adapters.PtrString("Default cluster parameter group for aurora-mysql5.7"),
			DBClusterParameterGroupArn:  adapters.PtrString("arn:aws:rds:eu-west-1:052392120703:cluster-pg:default.aurora-mysql5.7"),
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
				SupportedEngineModes: []string{
					"provisioned",
				},
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
				SupportedEngineModes: []string{
					"provisioned",
				},
			},
			{
				ParameterName: adapters.PtrString("aurora_binlog_replication_max_yield_seconds"),
				Description:   adapters.PtrString("Controls the number of seconds that binary log dump thread waits up to for the current binlog file to be filled by transactions. This wait period avoids contention that can arise from replicating each binlog event individually."),
				Source:        adapters.PtrString("engine-default"),
				ApplyType:     adapters.PtrString("dynamic"),
				DataType:      adapters.PtrString("integer"),
				AllowedValues: adapters.PtrString("0-36000"),
				IsModifiable:  adapters.PtrBool(true),
				ApplyMethod:   types.ApplyMethodPendingReboot,
				SupportedEngineModes: []string{
					"provisioned",
				},
			},
			{
				ParameterName: adapters.PtrString("aurora_enable_staggered_replica_restart"),
				Description:   adapters.PtrString("Allow Aurora replicas to follow a staggered restart schedule to increase cluster availability."),
				Source:        adapters.PtrString("system"),
				ApplyType:     adapters.PtrString("dynamic"),
				DataType:      adapters.PtrString("boolean"),
				AllowedValues: adapters.PtrString("0,1"),
				IsModifiable:  adapters.PtrBool(true),
				ApplyMethod:   types.ApplyMethodImmediate,
				SupportedEngineModes: []string{
					"provisioned",
				},
			},
		},
	}

	item, err := dBClusterParameterGroupItemMapper("", "foo", &group)

	if err != nil {
		t.Fatal(err)
	}

	if err = item.Validate(); err != nil {
		t.Error(err)
	}
}

func TestNewDBClusterParameterGroupSource(t *testing.T) {
	client, account, region := GetAutoConfig(t)

	source := NewDBClusterParameterGroupSource(client, account, region)

	test := adapters.E2ETest{
		Adapter: source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}

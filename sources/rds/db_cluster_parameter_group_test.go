package rds

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/overmindtech/aws-source/sources"
)

func TestDBClusterParameterGroupOutputMapper(t *testing.T) {
	group := ClusterParameterGroup{
		DBClusterParameterGroup: types.DBClusterParameterGroup{
			DBClusterParameterGroupName: sources.PtrString("default.aurora-mysql5.7"),
			DBParameterGroupFamily:      sources.PtrString("aurora-mysql5.7"),
			Description:                 sources.PtrString("Default cluster parameter group for aurora-mysql5.7"),
			DBClusterParameterGroupArn:  sources.PtrString("arn:aws:rds:eu-west-1:052392120703:cluster-pg:default.aurora-mysql5.7"),
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
				IsModifiable:   true,
				ApplyMethod:    types.ApplyMethodPendingReboot,
				SupportedEngineModes: []string{
					"provisioned",
				},
			},
			{
				ParameterName: sources.PtrString("allow-suspicious-udfs"),
				Description:   sources.PtrString("Controls whether user-defined functions that have only an xxx symbol for the main function can be loaded"),
				Source:        sources.PtrString("engine-default"),
				ApplyType:     sources.PtrString("static"),
				DataType:      sources.PtrString("boolean"),
				AllowedValues: sources.PtrString("0,1"),
				IsModifiable:  false,
				ApplyMethod:   types.ApplyMethodPendingReboot,
				SupportedEngineModes: []string{
					"provisioned",
				},
			},
			{
				ParameterName: sources.PtrString("aurora_binlog_replication_max_yield_seconds"),
				Description:   sources.PtrString("Controls the number of seconds that binary log dump thread waits up to for the current binlog file to be filled by transactions. This wait period avoids contention that can arise from replicating each binlog event individually."),
				Source:        sources.PtrString("engine-default"),
				ApplyType:     sources.PtrString("dynamic"),
				DataType:      sources.PtrString("integer"),
				AllowedValues: sources.PtrString("0-36000"),
				IsModifiable:  true,
				ApplyMethod:   types.ApplyMethodPendingReboot,
				SupportedEngineModes: []string{
					"provisioned",
				},
			},
			{
				ParameterName: sources.PtrString("aurora_enable_staggered_replica_restart"),
				Description:   sources.PtrString("Allow Aurora replicas to follow a staggered restart schedule to increase cluster availability."),
				Source:        sources.PtrString("system"),
				ApplyType:     sources.PtrString("dynamic"),
				DataType:      sources.PtrString("boolean"),
				AllowedValues: sources.PtrString("0,1"),
				IsModifiable:  true,
				ApplyMethod:   types.ApplyMethodImmediate,
				SupportedEngineModes: []string{
					"provisioned",
				},
			},
		},
	}

	item, err := dBClusterParameterGroupItemMapper("foo", &group)

	if err != nil {
		t.Fatal(err)
	}

	if err = item.Validate(); err != nil {
		t.Error(err)
	}
}

func TestNewDBClusterParameterGroupSource(t *testing.T) {
	config, account, region := sources.GetAutoConfig(t)

	source := NewDBClusterParameterGroupSource(config, account, region)

	test := sources.E2ETest{
		Source:  source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}

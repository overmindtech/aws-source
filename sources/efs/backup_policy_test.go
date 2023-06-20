package efs

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/efs"
	"github.com/aws/aws-sdk-go-v2/service/efs/types"
	"github.com/overmindtech/aws-source/sources"
)

func TestBackupPolicyOutputMapper(t *testing.T) {
	output := &efs.DescribeBackupPolicyOutput{
		BackupPolicy: &types.BackupPolicy{
			Status: types.StatusEnabled,
		},
	}

	items, err := BackupPolicyOutputMapper("foo", nil, output)

	if err != nil {
		t.Fatal(err)
	}

	for _, item := range items {
		if err := item.Validate(); err != nil {
			t.Error(err)
		}
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %v", len(items))
	}
}

func TestNewBackupPolicySource(t *testing.T) {
	config, account, _ := sources.GetAutoConfig(t)

	source := NewBackupPolicySource(config, account, &TestRateLimit)

	test := sources.E2ETest{
		Source:  source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}

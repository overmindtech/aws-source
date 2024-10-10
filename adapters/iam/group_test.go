package iam

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/overmindtech/aws-source/adapters"
)

func TestGroupItemMapper(t *testing.T) {
	zone := types.Group{
		Path:       adapters.PtrString("/"),
		GroupName:  adapters.PtrString("power-users"),
		GroupId:    adapters.PtrString("AGPA3VLV2U27T6SSLJMDS"),
		Arn:        adapters.PtrString("arn:aws:iam::801795385023:group/power-users"),
		CreateDate: adapters.PtrTime(time.Now()),
	}

	item, err := groupItemMapper("", "foo", &zone)

	if err != nil {
		t.Error(err)
	}

	if err = item.Validate(); err != nil {
		t.Error(err)
	}

}

func TestNewGroupSource(t *testing.T) {
	config, account, region := adapters.GetAutoConfig(t)
	client := iam.NewFromConfig(config, func(o *iam.Options) {
		o.RetryMode = aws.RetryModeAdaptive
		o.RetryMaxAttempts = 10
	})

	source := NewGroupSource(client, account, region)

	test := adapters.E2ETest{
		Adapter: source,
		Timeout: 30 * time.Second,
	}

	test.Run(t)
}

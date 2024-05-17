package iam

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/overmindtech/aws-source/sources"
)

func TestGroupItemMapper(t *testing.T) {
	zone := types.Group{
		Path:       sources.PtrString("/"),
		GroupName:  sources.PtrString("power-users"),
		GroupId:    sources.PtrString("AGPA3VLV2U27T6SSLJMDS"),
		Arn:        sources.PtrString("arn:aws:iam::801795385023:group/power-users"),
		CreateDate: sources.PtrTime(time.Now()),
	}

	item, err := groupItemMapper("foo", &zone)

	if err != nil {
		t.Error(err)
	}

	if err = item.Validate(); err != nil {
		t.Error(err)
	}

}

func TestNewGroupSource(t *testing.T) {
	config, account, region := sources.GetAutoConfig(t)

	source := NewGroupSource(config, account, region)

	test := sources.E2ETest{
		Source:  source,
		Timeout: 30 * time.Second,
	}

	test.Run(t)
}

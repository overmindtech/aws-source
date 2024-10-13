package adapters

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/overmindtech/aws-source/adapterhelpers"
)

func TestGroupItemMapper(t *testing.T) {
	zone := types.Group{
		Path:       adapterhelpers.PtrString("/"),
		GroupName:  adapterhelpers.PtrString("power-users"),
		GroupId:    adapterhelpers.PtrString("AGPA3VLV2U27T6SSLJMDS"),
		Arn:        adapterhelpers.PtrString("arn:aws:iam::801795385023:group/power-users"),
		CreateDate: adapterhelpers.PtrTime(time.Now()),
	}

	item, err := groupItemMapper("", "foo", &zone)

	if err != nil {
		t.Error(err)
	}

	if err = item.Validate(); err != nil {
		t.Error(err)
	}

}

func TestNewIAMGroupAdapter(t *testing.T) {
	config, account, region := adapterhelpers.GetAutoConfig(t)
	client := iam.NewFromConfig(config, func(o *iam.Options) {
		o.RetryMode = aws.RetryModeAdaptive
		o.RetryMaxAttempts = 10
	})

	adapter := NewIAMGroupAdapter(client, account, region)

	test := adapterhelpers.E2ETest{
		Adapter: adapter,
		Timeout: 30 * time.Second,
	}

	test.Run(t)
}

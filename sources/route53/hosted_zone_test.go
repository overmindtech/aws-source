package route53

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func TestHostedZoneItemMapper(t *testing.T) {
	zone := types.HostedZone{
		Id:              sources.PtrString("/hostedzone/Z08416862SZP5DJXIDB29"),
		Name:            sources.PtrString("overmind-demo.com."),
		CallerReference: sources.PtrString("RISWorkflow-RD:144d3779-1574-42bf-9e75-f309838ea0a1"),
		Config: &types.HostedZoneConfig{
			Comment:     sources.PtrString("HostedZone created by Route53 Registrar"),
			PrivateZone: false,
		},
		ResourceRecordSetCount: sources.PtrInt64(3),
		LinkedService: &types.LinkedService{
			Description:      sources.PtrString("service description"),
			ServicePrincipal: sources.PtrString("principal"),
		},
	}

	item, err := hostedZoneItemMapper("foo", &zone)

	if err != nil {
		t.Error(err)
	}

	if err = item.Validate(); err != nil {
		t.Error(err)
	}

	tests := sources.ItemRequestTests{
		{
			ExpectedType:   "route53-resource-record-set",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "/hostedzone/Z08416862SZP5DJXIDB29",
			ExpectedScope:  "foo",
		},
	}

	tests.Execute(t, item)
}

func TestNewHostedZoneSource(t *testing.T) {
	config, account, region := sources.GetAutoConfig(t)

	source := NewHostedZoneSource(config, account, region)

	test := sources.E2ETest{
		Source:  source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}

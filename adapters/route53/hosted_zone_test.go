package route53

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/overmindtech/aws-source/adapters"
	"github.com/overmindtech/sdp-go"
)

func TestHostedZoneItemMapper(t *testing.T) {
	zone := types.HostedZone{
		Id:              adapters.PtrString("/hostedzone/Z08416862SZP5DJXIDB29"),
		Name:            adapters.PtrString("overmind-demo.com."),
		CallerReference: adapters.PtrString("RISWorkflow-RD:144d3779-1574-42bf-9e75-f309838ea0a1"),
		Config: &types.HostedZoneConfig{
			Comment:     adapters.PtrString("HostedZone created by Route53 Registrar"),
			PrivateZone: false,
		},
		ResourceRecordSetCount: adapters.PtrInt64(3),
		LinkedService: &types.LinkedService{
			Description:      adapters.PtrString("service description"),
			ServicePrincipal: adapters.PtrString("principal"),
		},
	}

	item, err := hostedZoneItemMapper("", "foo", &zone)

	if err != nil {
		t.Error(err)
	}

	if err = item.Validate(); err != nil {
		t.Error(err)
	}

	tests := adapters.QueryTests{
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
	client, account, region := GetAutoConfig(t)

	source := NewHostedZoneSource(client, account, region)

	test := adapters.E2ETest{
		Adapter: source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}

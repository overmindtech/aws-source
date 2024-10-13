package lambda

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/lambda/types"

	"github.com/overmindtech/aws-source/adapterhelpers"
	"github.com/overmindtech/sdp-go"
)

func TestLayerItemMapper(t *testing.T) {
	layer := types.LayersListItem{
		LatestMatchingVersion: &types.LayerVersionsListItem{
			CompatibleArchitectures: []types.Architecture{
				types.ArchitectureArm64,
				types.ArchitectureX8664,
			},
			CompatibleRuntimes: []types.Runtime{
				types.RuntimeJava11,
			},
			CreatedDate:     adapterhelpers.PtrString("2018-11-27T15:10:45.123+0000"),
			Description:     adapterhelpers.PtrString("description"),
			LayerVersionArn: adapterhelpers.PtrString("arn:aws:service:region:account:type/id"),
			LicenseInfo:     adapterhelpers.PtrString("info"),
			Version:         10,
		},
		LayerArn:  adapterhelpers.PtrString("arn:aws:service:region:account:type/id"),
		LayerName: adapterhelpers.PtrString("name"),
	}

	item, err := layerItemMapper("", "foo", &layer)

	if err != nil {
		t.Error(err)
	}

	if err = item.Validate(); err != nil {
		t.Error(err)
	}

	tests := adapterhelpers.QueryTests{
		{
			ExpectedType:   "lambda-layer-version",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "name:10",
			ExpectedScope:  "foo",
		},
	}

	tests.Execute(t, item)
}

func TestNewLayerAdapter(t *testing.T) {
	client, account, region := GetAutoConfig(t)

	adapter := NewLayerAdapter(client, account, region)

	test := adapterhelpers.E2ETest{
		Adapter: adapter,
		Timeout: 10 * time.Second,
		SkipGet: true,
	}

	test.Run(t)
}

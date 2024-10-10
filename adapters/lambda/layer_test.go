package lambda

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/overmindtech/aws-source/adapters"
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
			CreatedDate:     adapters.PtrString("2018-11-27T15:10:45.123+0000"),
			Description:     adapters.PtrString("description"),
			LayerVersionArn: adapters.PtrString("arn:aws:service:region:account:type/id"),
			LicenseInfo:     adapters.PtrString("info"),
			Version:         10,
		},
		LayerArn:  adapters.PtrString("arn:aws:service:region:account:type/id"),
		LayerName: adapters.PtrString("name"),
	}

	item, err := layerItemMapper("", "foo", &layer)

	if err != nil {
		t.Error(err)
	}

	if err = item.Validate(); err != nil {
		t.Error(err)
	}

	tests := adapters.QueryTests{
		{
			ExpectedType:   "lambda-layer-version",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "name:10",
			ExpectedScope:  "foo",
		},
	}

	tests.Execute(t, item)
}

func TestNewLayerSource(t *testing.T) {
	client, account, region := GetAutoConfig(t)

	source := NewLayerSource(client, account, region)

	test := adapters.E2ETest{
		Adapter: source,
		Timeout: 10 * time.Second,
		SkipGet: true,
	}

	test.Run(t)
}

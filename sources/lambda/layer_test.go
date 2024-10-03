package lambda

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/overmindtech/aws-source/sources"
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
			CreatedDate:     sources.PtrString("2018-11-27T15:10:45.123+0000"),
			Description:     sources.PtrString("description"),
			LayerVersionArn: sources.PtrString("arn:aws:service:region:account:type/id"),
			LicenseInfo:     sources.PtrString("info"),
			Version:         10,
		},
		LayerArn:  sources.PtrString("arn:aws:service:region:account:type/id"),
		LayerName: sources.PtrString("name"),
	}

	item, err := layerItemMapper("", "foo", &layer)

	if err != nil {
		t.Error(err)
	}

	if err = item.Validate(); err != nil {
		t.Error(err)
	}

	tests := sources.QueryTests{
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

	test := sources.E2ETest{
		Adapter: source,
		Timeout: 10 * time.Second,
		SkipGet: true,
	}

	test.Run(t)
}

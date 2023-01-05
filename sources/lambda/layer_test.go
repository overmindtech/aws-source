package lambda

import (
	"testing"

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

	item, err := LayerItemMapper("foo", &layer)

	if err != nil {
		t.Error(err)
	}

	if err = item.Validate(); err != nil {
		t.Error(err)
	}

	tests := sources.ItemRequestTests{
		{
			ExpectedType:   "lambda-layer-version",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  "name:10",
			ExpectedScope:  "foo",
		},
	}

	tests.Execute(t, item)
}

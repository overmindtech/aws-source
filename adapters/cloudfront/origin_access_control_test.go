package cloudfront

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/overmindtech/aws-source/adapters"
)

func TestOriginAccessControlItemMapper(t *testing.T) {
	x := types.OriginAccessControl{
		Id: adapters.PtrString("test"),
		OriginAccessControlConfig: &types.OriginAccessControlConfig{
			Name:                          adapters.PtrString("example-name"),
			OriginAccessControlOriginType: types.OriginAccessControlOriginTypesS3,
			SigningBehavior:               types.OriginAccessControlSigningBehaviorsAlways,
			SigningProtocol:               types.OriginAccessControlSigningProtocolsSigv4,
			Description:                   adapters.PtrString("example-description"),
		},
	}

	item, err := originAccessControlItemMapper("", "test", &x)

	if err != nil {
		t.Fatal(err)
	}

	if err = item.Validate(); err != nil {
		t.Error(err)
	}
}

func TestNewOriginAccessControlSource(t *testing.T) {
	client, account, _ := GetAutoConfig(t)

	source := NewOriginAccessControlSource(client, account)

	test := adapters.E2ETest{
		Adapter: source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}

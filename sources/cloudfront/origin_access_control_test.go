package cloudfront

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/overmindtech/aws-source/sources"
)

func TestOriginAccessControlItemMapper(t *testing.T) {
	x := types.OriginAccessControl{
		Id: sources.PtrString("test"),
		OriginAccessControlConfig: &types.OriginAccessControlConfig{
			Name:                          sources.PtrString("example-name"),
			OriginAccessControlOriginType: types.OriginAccessControlOriginTypesS3,
			SigningBehavior:               types.OriginAccessControlSigningBehaviorsAlways,
			SigningProtocol:               types.OriginAccessControlSigningProtocolsSigv4,
			Description:                   sources.PtrString("example-description"),
		},
	}

	item, err := originAccessControlItemMapper("test", &x)

	if err != nil {
		t.Fatal(err)
	}

	if err = item.Validate(); err != nil {
		t.Error(err)
	}
}

func TestNewOriginAccessControlSource(t *testing.T) {
	config, account, _ := sources.GetAutoConfig(t)

	source := NewOriginAccessControlSource(config, account)

	test := sources.E2ETest{
		Source:  source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}

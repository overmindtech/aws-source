package networkmanager

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"github.com/overmindtech/sdp-go"
	"testing"
	"time"

	"github.com/overmindtech/aws-source/sources"
)

func (t *TestClient) GetSites(ctx context.Context, params *networkmanager.GetSitesInput, optFns ...func(*networkmanager.Options)) (*networkmanager.GetSitesOutput, error) {
	return &networkmanager.GetSitesOutput{
		Sites: []types.Site{
			{
				Tags:            []types.Tag{},
				SiteId:          sources.PtrString("site1"),
				GlobalNetworkId: sources.PtrString("default"),
			},
			{
				Tags:            []types.Tag{},
				SiteId:          sources.PtrString("site2"),
				GlobalNetworkId: sources.PtrString("other"),
			},
		},
	}, nil
}

func TestSiteSearchFunc(t *testing.T) {
	scope := "123456789012.eu-west-2"
	item, err := globalNetworkGetFunc(context.Background(), &TestClient{}, scope, &networkmanager.DescribeGlobalNetworksInput{})

	if err != nil {
		t.Error(err)
	}

	if err = item.Validate(); err != nil {
		t.Error(err)
	}

	tests := sources.QueryTests{
		{
			ExpectedType:   "networkmanager-site",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "default",
			ExpectedScope:  scope,
		},
	}

	tests.Execute(t, item)
}

func TestNewSite(t *testing.T) {
	config, account, _ := sources.GetAutoConfig(t)
	source := NewSiteSource(config, account, &TestRateLimit)
	test := sources.E2ETest{
		Source:   source,
		Timeout:  30 * time.Second,
		SkipList: true,
	}
	test.Run(t)
}

package eks

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/overmindtech/aws-source/sources"
)

var AddonTestClient = TestClient{
	DescribeAddonOutput: &eks.DescribeAddonOutput{
		Addon: &types.Addon{
			AddonName:           sources.PtrString("aws-ebs-csi-driver"),
			ClusterName:         sources.PtrString("dylan"),
			Status:              types.AddonStatusActive,
			AddonVersion:        sources.PtrString("v1.13.0-eksbuild.3"),
			ConfigurationValues: sources.PtrString("values"),
			MarketplaceInformation: &types.MarketplaceInformation{
				ProductId:  sources.PtrString("id"),
				ProductUrl: sources.PtrString("url"),
			},
			Publisher: sources.PtrString("publisher"),
			Owner:     sources.PtrString("owner"),
			Health: &types.AddonHealth{
				Issues: []types.AddonIssue{},
			},
			AddonArn:              sources.PtrString("arn:aws:eks:eu-west-2:801795385023:addon/dylan/aws-ebs-csi-driver/a2c29d0e-72c4-a702-7887-2f739f4fc189"),
			CreatedAt:             sources.PtrTime(time.Now()),
			ModifiedAt:            sources.PtrTime(time.Now()),
			ServiceAccountRoleArn: sources.PtrString("arn:aws:iam::801795385023:role/eks-csi-dylan"),
		},
	},
}

func TestAddonGetFunc(t *testing.T) {
	item, err := AddonGetFunc(context.Background(), AddonTestClient, "foo", &eks.DescribeAddonInput{})

	if err != nil {
		t.Error(err)
	}

	if err = item.Validate(); err != nil {
		t.Error(err)
	}
}

func TestNewAddonSource(t *testing.T) {
	config, account, region := sources.GetAutoConfig(t)

	source := NewAddonSource(config, account, region)

	test := sources.E2ETest{
		Source:            source,
		Timeout:           10 * time.Second,
		SkipNotFoundCheck: true,
	}

	test.Run(t)
}

package eks

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/overmindtech/aws-source/adapters"
)

var AddonTestClient = TestClient{
	DescribeAddonOutput: &eks.DescribeAddonOutput{
		Addon: &types.Addon{
			AddonName:           adapters.PtrString("aws-ebs-csi-driver"),
			ClusterName:         adapters.PtrString("dylan"),
			Status:              types.AddonStatusActive,
			AddonVersion:        adapters.PtrString("v1.13.0-eksbuild.3"),
			ConfigurationValues: adapters.PtrString("values"),
			MarketplaceInformation: &types.MarketplaceInformation{
				ProductId:  adapters.PtrString("id"),
				ProductUrl: adapters.PtrString("url"),
			},
			Publisher: adapters.PtrString("publisher"),
			Owner:     adapters.PtrString("owner"),
			Health: &types.AddonHealth{
				Issues: []types.AddonIssue{},
			},
			AddonArn:              adapters.PtrString("arn:aws:eks:eu-west-2:801795385023:addon/dylan/aws-ebs-csi-driver/a2c29d0e-72c4-a702-7887-2f739f4fc189"),
			CreatedAt:             adapters.PtrTime(time.Now()),
			ModifiedAt:            adapters.PtrTime(time.Now()),
			ServiceAccountRoleArn: adapters.PtrString("arn:aws:iam::801795385023:role/eks-csi-dylan"),
		},
	},
}

func TestAddonGetFunc(t *testing.T) {
	item, err := addonGetFunc(context.Background(), AddonTestClient, "foo", &eks.DescribeAddonInput{})

	if err != nil {
		t.Error(err)
	}

	if err = item.Validate(); err != nil {
		t.Error(err)
	}
}

func TestNewAddonSource(t *testing.T) {
	client, account, region := GetAutoConfig(t)

	source := NewAddonSource(client, account, region)

	test := adapters.E2ETest{
		Adapter:           source,
		Timeout:           10 * time.Second,
		SkipNotFoundCheck: true,
	}

	test.Run(t)
}

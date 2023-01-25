package eks

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

var NodeGroupClient = TestClient{
	DescribeNodegroupOutput: &eks.DescribeNodegroupOutput{
		Nodegroup: &types.Nodegroup{
			NodegroupName:  sources.PtrString("default-2022122213523169820000001f"),
			NodegroupArn:   sources.PtrString("arn:aws:eks:eu-west-2:801795385023:nodegroup/dylan/default-2022122213523169820000001f/98c29d0d-b22a-aaa3-445e-ebf71d43f67c"),
			ClusterName:    sources.PtrString("dylan"),
			Version:        sources.PtrString("1.24"),
			ReleaseVersion: sources.PtrString("1.24.7-20221112"),
			CreatedAt:      sources.PtrTime(time.Now()),
			ModifiedAt:     sources.PtrTime(time.Now()),
			Status:         types.NodegroupStatusActive,
			CapacityType:   types.CapacityTypesOnDemand,
			DiskSize:       sources.PtrInt32(100),
			RemoteAccess: &types.RemoteAccessConfig{
				Ec2SshKey: sources.PtrString("key"), // link
				SourceSecurityGroups: []string{
					"sg1", // link
				},
			},
			ScalingConfig: &types.NodegroupScalingConfig{
				MinSize:     sources.PtrInt32(1),
				MaxSize:     sources.PtrInt32(3),
				DesiredSize: sources.PtrInt32(1),
			},
			InstanceTypes: []string{
				"T3large",
			},
			Subnets: []string{
				"subnet0d1fabfe6794b5543", // link
			},
			AmiType:  types.AMITypesAl2Arm64,
			NodeRole: sources.PtrString("arn:aws:iam::801795385023:role/default-eks-node-group-20221222134106992000000003"),
			Labels:   map[string]string{},
			Taints: []types.Taint{
				{
					Effect: types.TaintEffectNoSchedule,
					Key:    sources.PtrString("key"),
					Value:  sources.PtrString("value"),
				},
			},
			Resources: &types.NodegroupResources{
				AutoScalingGroups: []types.AutoScalingGroup{
					{
						Name: sources.PtrString("eks-default-2022122213523169820000001f-98c29d0d-b22a-aaa3-445e-ebf71d43f67c"), // link
					},
				},
				RemoteAccessSecurityGroup: sources.PtrString("sg2"), // link
			},
			Health: &types.NodegroupHealth{
				Issues: []types.Issue{},
			},
			UpdateConfig: &types.NodegroupUpdateConfig{
				MaxUnavailablePercentage: sources.PtrInt32(33),
			},
			LaunchTemplate: &types.LaunchTemplateSpecification{
				Name:    sources.PtrString("default-2022122213523100410000001d"), // link
				Version: sources.PtrString("1"),
				Id:      sources.PtrString("lt-097e994ce7e14fcdc"),
			},
			Tags: map[string]string{},
		},
	},
}

func TestNodegroupGetFunc(t *testing.T) {
	item, err := NodegroupGetFunc(context.Background(), NodeGroupClient, "foo", &eks.DescribeNodegroupInput{})

	if err != nil {
		t.Error(err)
	}

	if err = item.Validate(); err != nil {
		t.Error(err)
	}

	// It doesn't really make sense to test anything other than the linked items
	// since the attributes are converted automatically
	tests := sources.ItemRequestTests{
		{
			ExpectedType:   "ec2-key-pair",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  "key",
			ExpectedScope:  item.Scope,
		},
		{
			ExpectedType:   "ec2-security-group",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  "sg1",
			ExpectedScope:  item.Scope,
		},
		{
			ExpectedType:   "ec2-subnet",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  "subnet0d1fabfe6794b5543",
			ExpectedScope:  item.Scope,
		},
		{
			ExpectedType:   "autoscaling-auto-scaling-group",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  "eks-default-2022122213523169820000001f-98c29d0d-b22a-aaa3-445e-ebf71d43f67c",
			ExpectedScope:  item.Scope,
		},
		{
			ExpectedType:   "ec2-security-group",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  "sg2",
			ExpectedScope:  item.Scope,
		},
		{
			ExpectedType:   "ec2-launch-template",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  "lt-097e994ce7e14fcdc",
			ExpectedScope:  item.Scope,
		},
	}

	tests.Execute(t, item)
}

func TestNewNodegroupSource(t *testing.T) {
	config, account, region := sources.GetAutoConfig(t)

	source := NewNodegroupSource(config, account, region)

	test := sources.E2ETest{
		Source:            source,
		Timeout:           10 * time.Second,
		SkipNotFoundCheck: true,
	}

	test.Run(t)
}

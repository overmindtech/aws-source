package iam

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/overmindtech/aws-source/sources"
)

func TestInstanceProfileItemMapper(t *testing.T) {
	profile := types.InstanceProfile{
		Arn:                 sources.PtrString("arn:aws:iam::123456789012:instance-profile/webserver"),
		CreateDate:          sources.PtrTime(time.Now()),
		InstanceProfileId:   sources.PtrString("AIDACKCEVSQ6C2EXAMPLE"),
		InstanceProfileName: sources.PtrString("webserver"),
		Path:                sources.PtrString("/"),
		Roles: []types.Role{
			{
				Arn:                      sources.PtrString("arn:aws:iam::123456789012:role/webserver"), // link
				CreateDate:               sources.PtrTime(time.Now()),
				Path:                     sources.PtrString("/"),
				RoleId:                   sources.PtrString("AIDACKCEVSQ6C2EXAMPLE"),
				RoleName:                 sources.PtrString("webserver"),
				AssumeRolePolicyDocument: sources.PtrString(`{}`),
				Description:              sources.PtrString("Allows EC2 instances to call AWS services on your behalf."),
				MaxSessionDuration:       sources.PtrInt32(3600),
				PermissionsBoundary: &types.AttachedPermissionsBoundary{
					PermissionsBoundaryArn:  sources.PtrString("arn:aws:iam::123456789012:policy/XCompanyBoundaries"), // link
					PermissionsBoundaryType: types.PermissionsBoundaryAttachmentTypePolicy,
				},
				RoleLastUsed: &types.RoleLastUsed{
					LastUsedDate: sources.PtrTime(time.Now()),
					Region:       sources.PtrString("us-east-1"),
				},
			},
		},
	}

	item, err := instanceProfileItemMapper("", "foo", &profile)

	if err != nil {
		t.Error(err)
	}

	if err = item.Validate(); err != nil {
		t.Error(err)
	}

}

func TestNewInstanceProfileSource(t *testing.T) {
	config, account, region := sources.GetAutoConfig(t)
	client := iam.NewFromConfig(config, func(o *iam.Options) {
		o.RetryMode = aws.RetryModeAdaptive
		o.RetryMaxAttempts = 10
	})

	source := NewInstanceProfileSource(client, account, region)

	test := sources.E2ETest{
		Adapter: source,
		Timeout: 30 * time.Second,
	}

	test.Run(t)
}

package iam

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/overmindtech/aws-source/adapters"
)

func TestInstanceProfileItemMapper(t *testing.T) {
	profile := types.InstanceProfile{
		Arn:                 adapters.PtrString("arn:aws:iam::123456789012:instance-profile/webserver"),
		CreateDate:          adapters.PtrTime(time.Now()),
		InstanceProfileId:   adapters.PtrString("AIDACKCEVSQ6C2EXAMPLE"),
		InstanceProfileName: adapters.PtrString("webserver"),
		Path:                adapters.PtrString("/"),
		Roles: []types.Role{
			{
				Arn:                      adapters.PtrString("arn:aws:iam::123456789012:role/webserver"), // link
				CreateDate:               adapters.PtrTime(time.Now()),
				Path:                     adapters.PtrString("/"),
				RoleId:                   adapters.PtrString("AIDACKCEVSQ6C2EXAMPLE"),
				RoleName:                 adapters.PtrString("webserver"),
				AssumeRolePolicyDocument: adapters.PtrString(`{}`),
				Description:              adapters.PtrString("Allows EC2 instances to call AWS services on your behalf."),
				MaxSessionDuration:       adapters.PtrInt32(3600),
				PermissionsBoundary: &types.AttachedPermissionsBoundary{
					PermissionsBoundaryArn:  adapters.PtrString("arn:aws:iam::123456789012:policy/XCompanyBoundaries"), // link
					PermissionsBoundaryType: types.PermissionsBoundaryAttachmentTypePolicy,
				},
				RoleLastUsed: &types.RoleLastUsed{
					LastUsedDate: adapters.PtrTime(time.Now()),
					Region:       adapters.PtrString("us-east-1"),
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
	config, account, region := adapters.GetAutoConfig(t)
	client := iam.NewFromConfig(config, func(o *iam.Options) {
		o.RetryMode = aws.RetryModeAdaptive
		o.RetryMaxAttempts = 10
	})

	source := NewInstanceProfileSource(client, account, region)

	test := adapters.E2ETest{
		Adapter: source,
		Timeout: 30 * time.Second,
	}

	test.Run(t)
}

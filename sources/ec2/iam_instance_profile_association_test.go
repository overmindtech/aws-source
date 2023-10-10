package ec2

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func TestIamInstanceProfileAssociationOutputMapper(t *testing.T) {
	output := ec2.DescribeIamInstanceProfileAssociationsOutput{
		IamInstanceProfileAssociations: []types.IamInstanceProfileAssociation{
			{
				AssociationId: sources.PtrString("eipassoc-1234567890abcdef0"),
				IamInstanceProfile: &types.IamInstanceProfile{
					Arn: sources.PtrString("arn:aws:iam::123456789012:instance-profile/webserver"), // link
					Id:  sources.PtrString("AIDACKCEVSQ6C2EXAMPLE"),
				},
				InstanceId: sources.PtrString("i-1234567890abcdef0"), // link
				State:      types.IamInstanceProfileAssociationStateAssociated,
				Timestamp:  sources.PtrTime(time.Now()),
			},
		},
	}

	items, err := iamInstanceProfileAssociationOutputMapper(context.Background(), nil, "foo", nil, &output)

	if err != nil {
		t.Error(err)
	}

	for _, item := range items {
		if err := item.Validate(); err != nil {
			t.Error(err)
		}
	}

	if len(items) != 1 {
		t.Errorf("expected 1 item, got %v", len(items))
	}

	item := items[0]

	// It doesn't really make sense to test anything other than the linked items
	// since the attributes are converted automatically
	tests := sources.QueryTests{
		{
			ExpectedType:   "iam-instance-profile",
			ExpectedQuery:  "arn:aws:iam::123456789012:instance-profile/webserver",
			ExpectedScope:  "123456789012",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
		},
		{
			ExpectedType:   "ec2-instance",
			ExpectedQuery:  "i-1234567890abcdef0",
			ExpectedScope:  "foo",
			ExpectedMethod: sdp.QueryMethod_GET,
		},
	}

	tests.Execute(t, item)
}

func TestNewIamInstanceProfileAssociationSource(t *testing.T) {
	config, account, _ := sources.GetAutoConfig(t)

	source := NewIamInstanceProfileAssociationSource(config, account, &TestRateLimit)

	test := sources.E2ETest{
		Source:  source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}

package ec2

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/overmindtech/aws-source/adapters"
	"github.com/overmindtech/sdp-go"
)

func TestIamInstanceProfileAssociationOutputMapper(t *testing.T) {
	output := ec2.DescribeIamInstanceProfileAssociationsOutput{
		IamInstanceProfileAssociations: []types.IamInstanceProfileAssociation{
			{
				AssociationId: adapters.PtrString("eipassoc-1234567890abcdef0"),
				IamInstanceProfile: &types.IamInstanceProfile{
					Arn: adapters.PtrString("arn:aws:iam::123456789012:instance-profile/webserver"), // link
					Id:  adapters.PtrString("AIDACKCEVSQ6C2EXAMPLE"),
				},
				InstanceId: adapters.PtrString("i-1234567890abcdef0"), // link
				State:      types.IamInstanceProfileAssociationStateAssociated,
				Timestamp:  adapters.PtrTime(time.Now()),
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
	tests := adapters.QueryTests{
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
	client, account, region := GetAutoConfig(t)

	source := NewIamInstanceProfileAssociationSource(client, account, region)

	test := adapters.E2ETest{
		Adapter: source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}

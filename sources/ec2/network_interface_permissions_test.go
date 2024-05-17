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

func TestNetworkInterfacePermissionInputMapperGet(t *testing.T) {
	input, err := networkInterfacePermissionInputMapperGet("foo", "bar")

	if err != nil {
		t.Error(err)
	}

	if len(input.NetworkInterfacePermissionIds) != 1 {
		t.Fatalf("expected 1 NetworkInterfacePermission ID, got %v", len(input.NetworkInterfacePermissionIds))
	}

	if input.NetworkInterfacePermissionIds[0] != "bar" {
		t.Errorf("expected NetworkInterfacePermission ID to be bar, got %v", input.NetworkInterfacePermissionIds[0])
	}
}

func TestNetworkInterfacePermissionInputMapperList(t *testing.T) {
	input, err := networkInterfacePermissionInputMapperList("foo")

	if err != nil {
		t.Error(err)
	}

	if len(input.Filters) != 0 || len(input.NetworkInterfacePermissionIds) != 0 {
		t.Errorf("non-empty input: %v", input)
	}
}

func TestNetworkInterfacePermissionOutputMapper(t *testing.T) {
	output := &ec2.DescribeNetworkInterfacePermissionsOutput{
		NetworkInterfacePermissions: []types.NetworkInterfacePermission{
			{
				NetworkInterfacePermissionId: sources.PtrString("eni-perm-0b6211455242c105e"),
				NetworkInterfaceId:           sources.PtrString("eni-07f8f3d404036c833"),
				AwsService:                   sources.PtrString("routing.hyperplane.eu-west-2.amazonaws.com"),
				Permission:                   types.InterfacePermissionTypeInstanceAttach,
				PermissionState: &types.NetworkInterfacePermissionState{
					State: types.NetworkInterfacePermissionStateCodeGranted,
				},
			},
		},
	}

	items, err := networkInterfacePermissionOutputMapper(context.Background(), nil, "foo", nil, output)

	if err != nil {
		t.Fatal(err)
	}

	for _, item := range items {
		if err := item.Validate(); err != nil {
			t.Error(err)
		}
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %v", len(items))
	}

	item := items[0]

	// It doesn't really make sense to test anything other than the linked items
	// since the attributes are converted automatically
	tests := sources.QueryTests{
		{
			ExpectedType:   "ec2-network-interface",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "eni-07f8f3d404036c833",
			ExpectedScope:  "foo",
		},
	}

	tests.Execute(t, item)

}

func TestNewNetworkInterfacePermissionSource(t *testing.T) {
	client, account, region := GetAutoConfig(t)

	source := NewNetworkInterfacePermissionSource(client, account, region)

	test := sources.E2ETest{
		Source:  source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}

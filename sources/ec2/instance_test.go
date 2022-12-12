package ec2

import (
	"context"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/overmindtech/sdp-go"
)

func checkItem(t *testing.T, item *sdp.ItemRequest, itemName string, expectedType string, expectedQuery string, expectedContext string) {
	if item.Type != expectedType {
		t.Errorf("%s.Type '%v' != '%v'", itemName, item.Type, expectedType)
	}
	if item.Method != sdp.RequestMethod_GET {
		t.Errorf("%s.Method '%v' != '%v'", itemName, item.Method, sdp.RequestMethod_GET)
	}
	if item.Query != expectedQuery {
		t.Errorf("%s.Query '%v' != '%v'", itemName, item.Query, expectedQuery)
	}
	if item.Context != expectedContext {
		t.Errorf("%s.Context '%v' != '%v'", itemName, item.Context, expectedContext)
	}
}

func TestInstanceMapping(t *testing.T) {
	t.Parallel()

	t.Run("empty", func(t *testing.T) {
		instance := types.Instance{}
		item, err := mapInstanceToItem(instance, "foo.bar")
		if err != nil {
			t.Error(err)
		}
		if item == nil {
			t.Error("item is nil")
		}
	})
	t.Run("with imageId", func(t *testing.T) {
		imageId := "imageId"
		instance := types.Instance{ImageId: &imageId}
		item, err := mapInstanceToItem(instance, "foo.bar")
		if err != nil {
			t.Error(err)
		}
		if item == nil {
			t.Fatal("item is nil")
		}
		if len(item.LinkedItemRequests) != 1 {
			t.Fatalf("unexpected LinkedItemRequests: %v", item)
		}
		checkItem(t, item.LinkedItemRequests[0], "image", "ec2-image", imageId, "foo.bar")
	})
	t.Run("with network interfaces", func(t *testing.T) {
		ipv6 := "2600::0"
		privateIp := "private ip"
		subnetId := "subnetId"
		vpcId := "vpcId"
		instance := types.Instance{
			NetworkInterfaces: []types.InstanceNetworkInterface{
				{
					Ipv6Addresses: []types.InstanceIpv6Address{{Ipv6Address: &ipv6}},
				},
				{
					PrivateIpAddresses: []types.InstancePrivateIpAddress{{PrivateIpAddress: &privateIp}},
				},
				{
					SubnetId: &subnetId,
				},
				{
					VpcId: &vpcId,
				},
			},
		}

		item, err := mapInstanceToItem(instance, "foo.bar")
		if err != nil {
			t.Fatal(err)
		}
		if item == nil {
			t.Fatal("item is nil")
		}
		if len(item.LinkedItemRequests) != 4 {
			t.Fatalf("unexpected LinkedItemRequests: %v", item)
		}
		checkItem(t, item.LinkedItemRequests[0], "ipv6Request", "ip", ipv6, "global")
		checkItem(t, item.LinkedItemRequests[1], "privateIpRequest", "ip", privateIp, "global")
		checkItem(t, item.LinkedItemRequests[2], "subnetRequest", "ec2-subnet", subnetId, "foo.bar")
		checkItem(t, item.LinkedItemRequests[3], "vpcRequest", "ec2-vpc", vpcId, "foo.bar")
	})
	t.Run("with public info", func(t *testing.T) {
		publicDns := "publicDns"
		publicIp := "publicIp"
		instance := types.Instance{
			PublicDnsName:   &publicDns,
			PublicIpAddress: &publicIp,
		}

		item, err := mapInstanceToItem(instance, "foo.bar")
		if err != nil {
			t.Fatal(err)
		}
		if item == nil {
			t.Fatal("item is nil")
		}
		if len(item.LinkedItemRequests) != 2 {
			t.Fatalf("unexpected LinkedItemRequests: %v", item)
		}
		checkItem(t, item.LinkedItemRequests[0], "publicDns", "dns", publicDns, "global")
		checkItem(t, item.LinkedItemRequests[1], "publicIp", "ip", publicIp, "global")
	})
}

type fakeClient func(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error)

func (m fakeClient) DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	return m(ctx, params, optFns...)
}

func createFakeClient(t *testing.T) fakeClient {
	clientCalls := 0
	return func(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
		clientCalls += 1
		if clientCalls > 2 {
			t.Error("Called fake client too often (>2)")
			return nil, nil
		}
		if params.NextToken == nil {
			nextToken := "page2"
			firstId := "first"
			return &ec2.DescribeInstancesOutput{
				NextToken: &nextToken,
				Reservations: []types.Reservation{
					{
						Instances: []types.Instance{
							{InstanceId: &firstId},
						},
					},
				},
			}, nil
		} else if *params.NextToken == "page2" {
			secondId := "second"
			return &ec2.DescribeInstancesOutput{
				Reservations: []types.Reservation{
					{
						Instances: []types.Instance{
							{InstanceId: &secondId},
						},
					},
				},
			}, nil
		}
		return nil, nil
	}
}

func TestGet(t *testing.T) {
	t.Parallel()
	t.Run("empty (context mismatch)", func(t *testing.T) {
		src := InstanceSource{}

		items, err := src.Get(context.Background(), "foo.bar", "query")
		if items != nil {
			t.Fatalf("unexpected items: %v", items)
		}
		if err == nil {
			t.Fatalf("expected err, got nil")
		}
		if !strings.HasPrefix(err.Error(), "requested context foo.bar does not match source context .") {
			t.Errorf("expected 'requested context foo.bar does not match source context .', got '%v'", err.Error())
		}
	})
	t.Run("with client", func(t *testing.T) {
		item, err := getImpl(context.Background(), createFakeClient(t), "foo.bar", "query")
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if item == nil {
			t.Fatalf("item is nil")
		}
		if item.Attributes.AttrStruct.Fields["instanceId"].GetStringValue() != "first" {
			t.Errorf("unexpected first item: %v", item)
		}
	})
}

func TestFind(t *testing.T) {
	t.Parallel()
	t.Run("empty (context mismatch)", func(t *testing.T) {
		src := InstanceSource{}

		items, err := src.Find(context.Background(), "foo.bar")
		if items != nil {
			t.Fatalf("unexpected items: %v", items)
		}
		if err == nil {
			t.Fatalf("expected err, got nil")
		}
		if !strings.HasPrefix(err.Error(), "requested context foo.bar does not match source context .") {
			t.Errorf("expected 'requested context foo.bar does not match source context .', got '%v'", err.Error())
		}
	})
	t.Run("with client", func(t *testing.T) {
		items, err := findImpl(context.Background(), createFakeClient(t), "foo.bar")
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if len(items) != 2 {
			t.Fatalf("unexpected items (len=%v): %v", len(items), items)
		}
		if items[0].Attributes.AttrStruct.Fields["instanceId"].GetStringValue() != "first" {
			t.Errorf("unexpected first item: %v", items[0])
		}
		if items[1].Attributes.AttrStruct.Fields["instanceId"].GetStringValue() != "second" {
			t.Errorf("unexpected second item: %v", items[0])
		}
	})
}
